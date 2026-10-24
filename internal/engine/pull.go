package engine

import (
	"fmt"
	"io"
	"maps"
	"strings"
	"time"

	"github.com/fabioueno/teire/internal/env"
	"github.com/fabioueno/teire/internal/fsx"
	"github.com/fabioueno/teire/internal/manifest"
	"github.com/fabioueno/teire/internal/state"
	"github.com/fabioueno/teire/internal/strategy"
)

// Pulled is one merge target that pull brought into the repository, with the
// owned paths it pulled.
type Pulled struct {
	Item  Item
	Paths []string
}

// Left is a target, or an owned path of it, that pull did not change.
type Left struct {
	Path   string
	Reason string
}

type PullResult struct {
	Pulled []Pulled
	Left   []Left
}

// Pull writes the owned merge values that drifted in the targets back into the
// repository. A linked target needs no pull, since its edits already land in
// the repository. With paths, it considers only those targets. On an error it
// still returns the state of the items it finished, so the caller saves it.
func Pull(e env.Env, p Plan, st state.State, paths []string, now time.Time) (state.State, PullResult, error) {
	next := state.State{Targets: maps.Clone(st.Targets)}
	if next.Targets == nil {
		next.Targets = map[string]state.Entry{}
	}

	items, err := selectItems(p, paths)
	if err != nil {
		return next, PullResult{}, err
	}

	var result PullResult

	backups := e.Backups(now)

	for _, item := range items {
		if item.Strategy == manifest.Merge {
			if err := pullMerge(item, next, &result, backups); err != nil {
				return next, result, fmt.Errorf("pull %s: %w", item.Path, err)
			}
		}

		if len(paths) > 0 && !touches(result, item.Path) {
			result.Left = append(result.Left, Left{Path: item.Path, Reason: "nothing to pull"})
		}
	}

	return next, result, nil
}

// selectItems gives every managed item, or the ones at paths. A path that is
// not a managed target stops the pull before it writes anything.
func selectItems(p Plan, paths []string) ([]Item, error) {
	var managed []Item

	for _, item := range p.Items {
		if !item.Removed {
			managed = append(managed, item)
		}
	}

	if len(paths) == 0 {
		return managed, nil
	}

	selected := make([]Item, 0, len(paths))

	for _, path := range paths {
		found := false

		for _, item := range managed {
			if item.Path == path {
				selected = append(selected, item)
				found = true

				break
			}
		}

		if !found {
			return nil, fmt.Errorf("pull %s: not a managed target", path)
		}
	}

	return selected, nil
}

func touches(r PullResult, path string) bool {
	for _, pulled := range r.Pulled {
		if pulled.Item.Path == path {
			return true
		}
	}

	for _, left := range r.Left {
		if left.Path == path {
			return true
		}
	}

	return false
}

func pullMerge(item Item, next state.State, result *PullResult, backups string) error {
	if item.Missing {
		return nil
	}

	var drifted []strategy.Change

	for _, c := range item.Merge {
		switch {
		case c.Action == strategy.Conflict:
			result.Left = append(result.Left, Left{Path: item.Path, Reason: c.Path + " changed in both"})
		case c.Action == strategy.Drift && c.Target == "":
			result.Left = append(result.Left, Left{Path: item.Path, Reason: c.Path + " is gone from the target"})
		case c.Action == strategy.Drift:
			drifted = append(drifted, c)
		}
	}

	if len(drifted) == 0 {
		return nil
	}

	if err := fsx.BackupCopy(item.Source, backups); err != nil {
		return err
	}

	if err := strategy.PullMerge(item.Source, drifted); err != nil {
		return err
	}

	next.Targets[item.Path] = state.Entry{Strategy: manifest.Merge, Owned: pulledOwned(item.Last.Owned, drifted)}

	paths := make([]string, 0, len(drifted))
	for _, c := range drifted {
		paths = append(paths, c.Path)
	}

	result.Pulled = append(result.Pulled, Pulled{Item: item, Paths: paths})

	return nil
}

// pulledOwned records the target value of each pulled path as the new last
// value, and keeps the rest of the record as it was.
func pulledOwned(last map[string]string, drifted []strategy.Change) map[string]string {
	settled := make([]strategy.Change, 0, len(drifted))
	for _, c := range drifted {
		settled = append(settled, strategy.Change{Path: c.Path, Action: strategy.Unchanged, Source: c.Target})
	}

	owned := maps.Clone(last)
	maps.Copy(owned, strategy.Owned(settled, nil))

	return owned
}

// RenderPull writes one line for each pulled target and each target left
// alone, then a summary line.
func RenderPull(w io.Writer, r PullResult, home string) error {
	var b strings.Builder

	for _, pulled := range r.Pulled {
		line(&b, "pull", ShortPath(pulled.Item.Path, home), strings.Join(pulled.Paths, ", "))
	}

	for _, left := range r.Left {
		line(&b, "left", ShortPath(left.Path, home), left.Reason)
	}

	if len(r.Pulled)+len(r.Left) == 0 {
		fmt.Fprintln(&b, "Nothing to pull.")
	} else {
		fmt.Fprintf(&b, "%d pulled, %d left.\n", len(r.Pulled), len(r.Left))
	}

	_, err := io.WriteString(w, b.String())

	return err
}

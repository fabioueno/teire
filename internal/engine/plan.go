// Package engine turns resolved targets and the last apply into a plan, and
// carries the plan out.
package engine

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"

	"github.com/fabioueno/teire/internal/env"
	"github.com/fabioueno/teire/internal/manifest"
	"github.com/fabioueno/teire/internal/state"
	"github.com/fabioueno/teire/internal/strategy"
)

// Item is the plan for one target. Link is set for link and folder targets,
// Merge for merge targets. An item with Removed set has left the manifest.
type Item struct {
	Package  string
	Path     string
	Strategy manifest.Strategy
	Source   string
	Removed  bool
	Missing  bool
	Link     strategy.LinkChange
	Merge    []strategy.Change
	Last     state.Entry
}

type Plan struct {
	Items []Item
}

func Build(e env.Env, targets []manifest.Target, st state.State) (Plan, error) {
	items := make([]Item, 0, len(targets))
	planned := make(map[string]bool, len(targets))

	for _, t := range targets {
		item, err := planTarget(e, t, st.Targets[t.Path])
		if err != nil {
			return Plan{}, fmt.Errorf("plan %s: %w", t.Path, err)
		}

		items = append(items, item)
		planned[t.Path] = true
	}

	for _, path := range slices.Sorted(maps.Keys(st.Targets)) {
		if planned[path] {
			continue
		}

		item, err := planRemoved(path, st.Targets[path])
		if err != nil {
			return Plan{}, fmt.Errorf("plan %s: %w", path, err)
		}

		items = append(items, item)
	}

	return Plan{Items: items}, nil
}

func planTarget(e env.Env, t manifest.Target, last state.Entry) (Item, error) {
	item := Item{
		Package:  t.Package,
		Path:     t.Path,
		Strategy: t.Strategy,
		Source:   filepath.Join(e.Repo, filepath.FromSlash(t.Source)),
		Last:     last,
	}

	switch t.Strategy {
	case manifest.Link, manifest.Folder:
		c, err := strategy.PlanLink(item.Path, item.Source, e.Repo, t.Strategy == manifest.Folder)
		if err != nil {
			return Item{}, err
		}

		item.Link = c
	case manifest.Merge:
		source, err := os.ReadFile(item.Source)
		if err != nil {
			return Item{}, fmt.Errorf("read source: %w", err)
		}

		target, err := os.ReadFile(item.Path)
		if errors.Is(err, fs.ErrNotExist) {
			item.Missing = true
			return item, nil
		}

		if err != nil {
			return Item{}, fmt.Errorf("read target: %w", err)
		}

		changes, err := strategy.PlanMerge(source, target, t.Own, last.Owned)
		if err != nil {
			return Item{}, err
		}

		item.Merge = changes
	default:
		return Item{}, fmt.Errorf("unknown strategy %q", t.Strategy)
	}

	return item, nil
}

func planRemoved(path string, last state.Entry) (Item, error) {
	item := Item{Path: path, Strategy: last.Strategy, Source: last.Source, Removed: true, Last: last}

	switch last.Strategy {
	case manifest.Link, manifest.Folder:
		item.Link = strategy.PlanUnlink(path, last.Source)
	case manifest.Merge:
		target, err := os.ReadFile(path)
		if errors.Is(err, fs.ErrNotExist) {
			item.Missing = true
			return item, nil
		}

		if err != nil {
			return Item{}, fmt.Errorf("read target: %w", err)
		}

		changes, err := strategy.PlanMerge([]byte("{}"), target, nil, last.Owned)
		if err != nil {
			return Item{}, err
		}

		item.Merge = changes
	default:
		return Item{}, fmt.Errorf("unknown strategy %q in state", last.Strategy)
	}

	return item, nil
}

// Changed reports whether the target differs from what the manifest asks for.
func (i Item) Changed() bool {
	if i.Removed || i.Missing {
		return true
	}

	switch i.Strategy {
	case manifest.Link, manifest.Folder:
		return i.Link.Action != strategy.LinkUnchanged
	case manifest.Merge:
		for _, c := range i.Merge {
			if c.Action != strategy.Unchanged {
				return true
			}
		}
	}

	return false
}

func (i Item) Conflicts() int {
	switch i.Strategy {
	case manifest.Link, manifest.Folder:
		if i.Link.Action == strategy.LinkConflict {
			return 1
		}
	case manifest.Merge:
		n := 0
		for _, c := range i.Merge {
			if c.Action == strategy.Conflict {
				n++
			}
		}

		return n
	}

	return 0
}

func (p Plan) Changed() bool {
	return slices.ContainsFunc(p.Items, Item.Changed)
}

func (p Plan) Conflicts() int {
	n := 0
	for _, item := range p.Items {
		n += item.Conflicts()
	}

	return n
}

package engine

import (
	"fmt"
	"maps"
	"slices"
	"time"

	"github.com/fabioueno/teire/internal/env"
	"github.com/fabioueno/teire/internal/fsx"
	"github.com/fabioueno/teire/internal/manifest"
	"github.com/fabioueno/teire/internal/state"
	"github.com/fabioueno/teire/internal/strategy"
)

type Choice int

const (
	Skip Choice = iota
	Overwrite
)

// Conflict is one question for the person running apply. Path is the owned
// path for a merge, and empty for a link.
type Conflict struct {
	Item   Item
	Path   string
	Reason string
	Diff   string
}

type Resolver func(Conflict) (Choice, error)

type Result struct {
	Skipped int
	Missing []string
	NotKept map[string][]string
}

// Apply carries out the plan and returns the state to save. On an error it
// still returns the state of the items it finished, so the caller saves it.
func Apply(e env.Env, p Plan, st state.State, resolve Resolver, now time.Time) (state.State, Result, error) {
	next := state.State{Targets: maps.Clone(st.Targets)}
	if next.Targets == nil {
		next.Targets = map[string]state.Entry{}
	}

	result := Result{NotKept: map[string][]string{}}
	backups := e.Backups(now)

	for _, item := range p.Items {
		var err error

		switch item.Strategy {
		case manifest.Files, manifest.Folder:
			err = applyLink(item, next, &result, resolve, backups)
		case manifest.Merge:
			err = applyMerge(item, next, &result, resolve, backups)
		default:
			err = fmt.Errorf("unknown strategy %q", item.Strategy)
		}

		if err != nil {
			return next, result, fmt.Errorf("apply %s: %w", item.Path, err)
		}
	}

	return next, result, nil
}

func applyLink(item Item, next state.State, result *Result, resolve Resolver, backups string) error {
	c := item.Link
	if c.Action == strategy.LinkConflict {
		diff, err := linkDiff(c)
		if err != nil {
			return err
		}

		choice, err := resolve(Conflict{Item: item, Reason: c.Reason, Diff: diff})
		if err != nil {
			return err
		}

		if choice != Overwrite {
			result.Skipped++
			return nil
		}

		c.Action = strategy.LinkReplace
	}

	if err := strategy.ApplyLink(c, backups); err != nil {
		return err
	}

	switch c.Action {
	case strategy.LinkCreate, strategy.LinkReplace, strategy.LinkUnchanged:
		next.Targets[item.Path] = state.Entry{Strategy: item.Strategy, Source: c.Source}
	case strategy.LinkRemove, strategy.LinkRelease:
		delete(next.Targets, item.Path)
	case strategy.LinkConflict:
	}

	return nil
}

func applyMerge(item Item, next state.State, result *Result, resolve Resolver, backups string) error {
	if item.Missing {
		if item.Removed {
			delete(next.Targets, item.Path)
		} else {
			result.Missing = append(result.Missing, item.Path)
		}

		return nil
	}

	changes := slices.Clone(item.Merge)
	for i, c := range changes {
		if c.Action != strategy.Conflict {
			continue
		}

		choice, err := resolve(Conflict{Item: item, Path: c.Path, Reason: "both changed", Diff: valueDiff(c)})
		if err != nil {
			return err
		}

		if choice != Overwrite {
			result.Skipped++
			continue
		}

		changes[i].Action = strategy.Set
	}

	if writes(changes) {
		if err := fsx.BackupCopy(item.Path, backups); err != nil {
			return err
		}

		notKept, err := strategy.ApplyMerge(item.Path, changes)
		if err != nil {
			return err
		}

		if len(notKept) > 0 {
			result.NotKept[item.Path] = notKept
		}
	}

	owned := strategy.Owned(changes, item.Last.Owned)
	if len(owned) == 0 {
		delete(next.Targets, item.Path)
		return nil
	}

	next.Targets[item.Path] = state.Entry{Strategy: manifest.Merge, Owned: owned}

	return nil
}

func writes(changes []strategy.Change) bool {
	return slices.ContainsFunc(changes, func(c strategy.Change) bool {
		return c.Action == strategy.Set || c.Action == strategy.Delete
	})
}

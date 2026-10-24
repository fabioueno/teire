// Package strategy plans and applies the ways teire writes a target.
package strategy

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/tidwall/gjson"
)

type Action string

const (
	Unchanged Action = "unchanged"
	Set       Action = "set"
	Delete    Action = "delete"
	Release   Action = "release"
	Drift     Action = "drift"
	Conflict  Action = "conflict"
)

// Change is the plan for one owned path. Source and Target hold the JSON value
// at the path, without comments, or "" when the path is absent.
type Change struct {
	Path   string
	Action Action
	Source string
	Target string
}

func PlanMerge(source, target []byte, own []string, last map[string]string) ([]Change, error) {
	if !validDocument(source) {
		return nil, errors.New("source is not valid JSON or JSONC")
	}

	if !validDocument(target) {
		return nil, errors.New("target is not valid JSON or JSONC")
	}

	changes := make([]Change, 0, len(own)+len(last))
	for _, path := range own {
		if err := checkPath(path); err != nil {
			return nil, err
		}

		src := lookup(source, path)
		if !src.Exists() {
			return nil, fmt.Errorf("owned path %q is not in the source", path)
		}

		lastHash, known := last[path]
		changes = append(changes, planOwned(path, src, lookup(target, path), lastHash, known))
	}

	for _, path := range slices.Sorted(maps.Keys(last)) {
		if slices.Contains(own, path) {
			continue
		}

		changes = append(changes, planReleased(path, lookup(target, path), last[path]))
	}

	return changes, nil
}

func planOwned(path string, src, tgt gjson.Result, lastHash string, known bool) Change {
	c := Change{Path: path, Source: src.Raw, Target: tgt.Raw}
	srcHash, tgtHash := hashValue(src), hashValue(tgt)

	switch {
	case srcHash == tgtHash:
		c.Action = Unchanged
	case !known && !tgt.Exists():
		c.Action = Set
	case !known:
		c.Action = Conflict
	case tgtHash == lastHash:
		c.Action = Set
	case srcHash == lastHash:
		c.Action = Drift
	default:
		c.Action = Conflict
	}

	return c
}

func planReleased(path string, tgt gjson.Result, lastHash string) Change {
	c := Change{Path: path, Action: Release, Target: tgt.Raw}
	if tgt.Exists() && hashValue(tgt) == lastHash {
		c.Action = Delete
	}

	return c
}

// Owned gives the owned paths and value hashes to record after the changes
// are applied. A path in drift or conflict keeps its last hash, so the next
// plan finds it again.
func Owned(changes []Change, last map[string]string) map[string]string {
	owned := map[string]string{}
	for _, c := range changes {
		switch c.Action {
		case Unchanged, Set:
			owned[c.Path] = hashRaw(c.Source)
		case Drift, Conflict:
			if lastHash, ok := last[c.Path]; ok {
				owned[c.Path] = lastHash
			}
		case Delete, Release:
		}
	}

	return owned
}

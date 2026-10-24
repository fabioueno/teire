package strategy

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/fabioueno/teire/internal/fsx"
)

var ErrTargetMissing = errors.New("target does not exist")

// MergeResult lists the owned paths that a merge did not settle. Moved paths
// changed in the target after the plan, so the merge left them alone. NotKept
// paths were written but read back with another value.
type MergeResult struct {
	Moved   []string
	NotKept []string
}

// ApplyMerge writes the set and delete changes into the target file. It reads
// the target again first, so the keys another program wrote since the plan
// stay as they are, and an owned value that changed since the plan is left
// for the next plan to show.
func ApplyMerge(path string, changes []Change) (MergeResult, error) {
	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return MergeResult{}, fmt.Errorf("%s: %w", path, ErrTargetMissing)
	}

	if err != nil {
		return MergeResult{}, fmt.Errorf("stat %s: %w", path, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return MergeResult{}, fmt.Errorf("read %s: %w", path, err)
	}

	applied, moved := splitMoved(data, changes)

	out, err := applyChanges(data, applied)
	if err != nil {
		return MergeResult{}, fmt.Errorf("merge into %s: %w", path, err)
	}

	if !bytes.Equal(out, data) {
		if err := fsx.WriteFileAtomic(path, out, info.Mode().Perm()); err != nil {
			return MergeResult{}, err
		}
	}

	written, err := os.ReadFile(path)
	if err != nil {
		return MergeResult{}, fmt.Errorf("read %s again: %w", path, err)
	}

	return MergeResult{Moved: moved, NotKept: notKept(written, applied)}, nil
}

// splitMoved separates the writes whose target value is still the one the
// plan saw from those whose value changed since.
func splitMoved(data []byte, changes []Change) ([]Change, []string) {
	applied := make([]Change, 0, len(changes))
	var moved []string

	for _, c := range changes {
		switch c.Action {
		case Set, Delete:
			if hashValue(lookup(data, c.Path)) != plannedHash(c) {
				moved = append(moved, c.Path)
				continue
			}
		case Unchanged, Release, Drift, Conflict:
		}

		applied = append(applied, c)
	}

	return applied, moved
}

func plannedHash(c Change) string {
	if c.Target == "" {
		return ""
	}

	return hashRaw(c.Target)
}

func applyChanges(data []byte, changes []Change) ([]byte, error) {
	out := data
	for _, c := range changes {
		var err error

		switch c.Action {
		case Set:
			out, err = setPath(out, c.Path, compact(c.Source))
		case Delete:
			out, err = deletePath(out, c.Path)
		case Unchanged, Release, Drift, Conflict:
		}

		if err != nil {
			return nil, fmt.Errorf("path %q: %w", c.Path, err)
		}
	}

	return out, nil
}

func notKept(written []byte, changes []Change) []string {
	var paths []string
	for _, c := range changes {
		value := lookup(written, c.Path)

		switch c.Action {
		case Set:
			if hashValue(value) != hashRaw(c.Source) {
				paths = append(paths, c.Path)
			}
		case Delete:
			if value.Exists() {
				paths = append(paths, c.Path)
			}
		case Unchanged, Release, Drift, Conflict:
		}
	}

	return paths
}

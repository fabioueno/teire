package strategy

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"github.com/fabioueno/teire/internal/fsx"
)

var ErrTargetMissing = errors.New("target does not exist")

// ApplyMerge writes the set and delete changes into the target file and
// returns the owned paths whose value did not stay after the write. It reads
// the target again first, so the keys another program wrote since the plan
// stay as they are.
func ApplyMerge(path string, changes []Change) ([]string, error) {
	info, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%s: %w", path, ErrTargetMissing)
	}

	if err != nil {
		return nil, fmt.Errorf("stat %s: %w", path, err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	out, err := applyChanges(data, changes)
	if err != nil {
		return nil, fmt.Errorf("merge into %s: %w", path, err)
	}

	if !bytes.Equal(out, data) {
		if err := fsx.WriteFileAtomic(path, out, info.Mode().Perm()); err != nil {
			return nil, err
		}
	}

	written, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s again: %w", path, err)
	}

	return notKept(written, changes), nil
}

func applyChanges(data []byte, changes []Change) ([]byte, error) {
	out := data
	for _, c := range changes {
		var err error

		switch c.Action {
		case Set:
			out, err = sjson.SetRawBytes(out, c.Path, compact(c.Source))
		case Delete:
			out, err = sjson.DeleteBytes(out, c.Path)
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
		value := gjson.GetBytes(written, c.Path)

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

// Package state keeps the record of the last apply, so that teire can tell a
// change in the repository from a change in the target.
package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/fabioueno/teire/internal/fsx"
	"github.com/fabioueno/teire/internal/manifest"
)

const version = 1

type State struct {
	Targets map[string]Entry
}

type Entry struct {
	Strategy manifest.Strategy `json:"strategy"`
	Owned    map[string]string `json:"owned,omitempty"`
}

type file struct {
	Version int              `json:"version"`
	Targets map[string]Entry `json:"targets"`
}

func Load(path string) (State, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return State{Targets: map[string]Entry{}}, nil
	}

	if err != nil {
		return State{}, fmt.Errorf("read state: %w", err)
	}

	var f file
	if err := json.Unmarshal(data, &f); err != nil {
		return State{}, fmt.Errorf("parse state %s: %w", path, err)
	}

	if f.Version != version {
		return State{}, fmt.Errorf("state %s has version %d, want %d", path, f.Version, version)
	}

	if f.Targets == nil {
		f.Targets = map[string]Entry{}
	}

	return State{Targets: f.Targets}, nil
}

func (s State) Save(path string) error {
	data, err := json.MarshalIndent(file{Version: version, Targets: s.Targets}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode state: %w", err)
	}

	if err := fsx.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create state folder: %w", err)
	}

	return fsx.WriteFileAtomic(path, append(data, '\n'), 0o600)
}

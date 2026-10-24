// Package env collects the folders teire works in. It is the only package that
// reads the process environment, so the others can run against any folders.
package env

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// BackupLayout is the time layout of a backup folder name. It sorts by time.
const BackupLayout = "20060102T150405Z"

type Env struct {
	Home      string
	Repo      string
	StateDir  string
	ConfigDir string
}

func FromOS() (Env, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Env{}, fmt.Errorf("find home folder: %w", err)
	}

	return New(home, os.Getenv("XDG_STATE_HOME"), os.Getenv("XDG_CONFIG_HOME")), nil
}

// New applies the XDG defaults when stateHome or configHome is empty.
func New(home, stateHome, configHome string) Env {
	if stateHome == "" {
		stateHome = filepath.Join(home, ".local", "state")
	}

	if configHome == "" {
		configHome = filepath.Join(home, ".config")
	}

	return Env{
		Home:      home,
		StateDir:  filepath.Join(stateHome, "teire"),
		ConfigDir: filepath.Join(configHome, "teire"),
	}
}

func (e Env) StatePath() string {
	return filepath.Join(e.StateDir, "state.json")
}

func (e Env) HostFile() string {
	return filepath.Join(e.ConfigDir, "host")
}

func (e Env) BackupsDir() string {
	return filepath.Join(e.StateDir, "backups")
}

func (e Env) Backups(now time.Time) string {
	return filepath.Join(e.BackupsDir(), now.UTC().Format(BackupLayout))
}

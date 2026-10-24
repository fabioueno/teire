package cli

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/fabioueno/teire/internal/engine"
	"github.com/fabioueno/teire/internal/env"
	"github.com/fabioueno/teire/internal/host"
	"github.com/fabioueno/teire/internal/manifest"
	"github.com/fabioueno/teire/internal/state"
)

const manifestName = "teire.toml"

type workspace struct {
	env      env.Env
	manifest manifest.Manifest
	host     string
	state    state.State
}

func load(d deps, opts *options) (workspace, error) {
	e, err := d.env()
	if err != nil {
		return workspace{}, err
	}

	e.Repo, err = findRepo(opts.repo, d.getwd)
	if err != nil {
		return workspace{}, err
	}

	m, err := readManifest(filepath.Join(e.Repo, manifestName))
	if err != nil {
		return workspace{}, err
	}

	name, err := host.Name(opts.host, e.HostFile(), d.localHostName)
	if err != nil {
		return workspace{}, err
	}

	st, err := state.Load(e.StatePath())
	if err != nil {
		return workspace{}, err
	}

	return workspace{env: e, manifest: m, host: name, state: st}, nil
}

func (w workspace) plan() (engine.Plan, error) {
	targets, err := manifest.Resolve(w.manifest, w.host, os.DirFS(w.env.Repo), w.env.Home)
	if err != nil {
		return engine.Plan{}, err
	}

	return engine.Build(w.env, targets, w.state)
}

func readManifest(path string) (manifest.Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return manifest.Manifest{}, fmt.Errorf("open manifest: %w", err)
	}

	defer func() { _ = f.Close() }()

	return manifest.Parse(f)
}

// findRepo gives the repository's real path, without links. Links record the
// path teire saw, so reaching the same repository through a link such as
// ~/.dotfiles must not give a second set of link paths.
func findRepo(flag string, getwd func() (string, error)) (string, error) {
	repo, err := locateRepo(flag, getwd)
	if err != nil {
		return "", err
	}

	return filepath.EvalSymlinks(repo)
}

// locateRepo takes the --repo flag, or else the nearest folder from the
// current one upward that holds a manifest.
func locateRepo(flag string, getwd func() (string, error)) (string, error) {
	if flag != "" {
		repo, err := filepath.Abs(flag)
		if err != nil {
			return "", err
		}

		if _, err := os.Stat(filepath.Join(repo, manifestName)); err != nil {
			return "", fmt.Errorf("repository %s: %w", repo, err)
		}

		return repo, nil
	}

	dir, err := getwd()
	if err != nil {
		return "", err
	}

	for {
		_, err := os.Stat(filepath.Join(dir, manifestName))
		if err == nil {
			return dir, nil
		}

		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no %s in this folder or above; pass --repo", manifestName)
		}

		dir = parent
	}
}

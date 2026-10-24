package engine

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/fabioueno/teire/internal/env"
	"github.com/fabioueno/teire/internal/manifest"
	"github.com/fabioueno/teire/internal/state"
)

var now = time.Date(2026, 10, 24, 10, 15, 0, 0, time.UTC)

// world is a copy of testdata/repo and testdata/home in a temporary folder.
type world struct {
	t   *testing.T
	env env.Env
}

func newWorld(t *testing.T) world {
	t.Helper()

	root := t.TempDir()
	e := env.New(filepath.Join(root, "home"), filepath.Join(root, "state"), filepath.Join(root, "config"))
	e.Repo = filepath.Join(root, "repo")

	require.NoError(t, os.CopyFS(e.Repo, os.DirFS(filepath.Join("testdata", "repo"))))
	require.NoError(t, os.CopyFS(e.Home, os.DirFS(filepath.Join("testdata", "home"))))

	return world{t: t, env: e}
}

func (w world) home(rel string) string {
	return filepath.Join(w.env.Home, rel)
}

func (w world) repo(rel string) string {
	return filepath.Join(w.env.Repo, rel)
}

// targets gives the two targets every engine test plans: a link and a merge.
func (w world) targets() []manifest.Target {
	return []manifest.Target{
		{Package: "claude", Source: "claude/claude.json", Path: w.home(".claude.json"), Strategy: manifest.Merge, Own: []string{"mcpServers.notes"}},
		{Package: "zsh", Source: "zsh/.zshrc", Path: w.home(".zshrc"), Strategy: manifest.Files},
	}
}

func (w world) copyFile(from, to string) {
	w.t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", from))
	require.NoError(w.t, err)
	require.NoError(w.t, os.MkdirAll(filepath.Dir(to), 0o755))
	require.NoError(w.t, os.WriteFile(to, data, 0o644))
}

func (w world) read(path string) string {
	w.t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(w.t, err)

	return string(data)
}

func emptyState() state.State {
	return state.State{Targets: map[string]state.Entry{}}
}

func golden(t *testing.T, name string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", "golden", name))
	require.NoError(t, err)

	return string(data)
}

// applied applies the targets to the world and gives the state it recorded.
func (w world) applied() state.State {
	w.t.Helper()

	p, err := Build(w.env, w.targets(), emptyState())
	require.NoError(w.t, err)

	st, _, err := Apply(w.env, p, emptyState(), func(Conflict) (Choice, error) { return Skip, nil }, now)
	require.NoError(w.t, err)

	return st
}

// replaceLink swaps the link at path for a plain file, the way a program does
// when it writes a temporary file and renames it over the link.
func (w world) replaceLink(path, fixture string) {
	w.t.Helper()

	require.NoError(w.t, os.Remove(path))
	w.copyFile(fixture, path)
}

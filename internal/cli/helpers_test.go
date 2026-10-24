package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/fabioueno/teire/internal/env"
)

var now = time.Date(2026, 10, 24, 10, 15, 0, 0, time.UTC)

// harness runs teire commands against a copy of testdata/repo and
// testdata/home in a temporary folder.
type harness struct {
	t        *testing.T
	root     string
	repo     string
	home     string
	wd       string
	stdin    string
	terminal bool
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	// The repository path teire records is the real one, and macOS puts
	// temporary folders below a link (/var -> /private/var).
	root, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)

	h := &harness{
		t:    t,
		root: root,
		repo: filepath.Join(root, "repo"),
		home: filepath.Join(root, "home"),
	}
	h.wd = h.repo

	require.NoError(t, os.CopyFS(h.repo, os.DirFS(filepath.Join("testdata", "repo"))))
	require.NoError(t, os.CopyFS(h.home, os.DirFS(filepath.Join("testdata", "home"))))

	return h
}

func (h *harness) env() env.Env {
	return env.New(h.home, filepath.Join(h.root, "state"), filepath.Join(h.root, "config"))
}

func (h *harness) run(args ...string) (string, error) {
	h.t.Helper()

	d := deps{
		env:           func() (env.Env, error) { return h.env(), nil },
		getwd:         func() (string, error) { return h.wd, nil },
		localHostName: func() (string, error) { return "MacBook-Pro", nil },
		stdin:         strings.NewReader(h.stdin),
		isTerminal:    func() bool { return h.terminal },
		now:           func() time.Time { return now },
	}

	root := newRoot(d)

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)

	err := root.Execute()

	return out.String(), err
}

// copyFile puts a file from testdata into the temporary folder.
func (h *harness) copyFile(from, to string) {
	h.t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", from))
	require.NoError(h.t, err)
	require.NoError(h.t, os.MkdirAll(filepath.Dir(to), 0o755))
	require.NoError(h.t, os.WriteFile(to, data, 0o644))
}

// replaceLink swaps the link for a plain file, the way a program does when it
// writes a temporary file and renames it over the link.
func (h *harness) replaceLink(path, fixture string) {
	h.t.Helper()

	require.NoError(h.t, os.Remove(path))
	h.copyFile(fixture, path)
}

func (h *harness) read(path string) string {
	h.t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(h.t, err)

	return string(data)
}

func (h *harness) readlink(path string) string {
	h.t.Helper()

	dest, err := os.Readlink(path)
	require.NoError(h.t, err)

	return dest
}

func golden(t *testing.T, name string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", "golden", name))
	require.NoError(t, err)

	return string(data)
}

func exitCode(t *testing.T, err error) int {
	t.Helper()

	if err == nil {
		return 0
	}

	code, _ := ExitCode(err)

	return code
}

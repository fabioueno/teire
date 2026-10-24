package engine

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fabioueno/teire/internal/manifest"
)

func TestPullMergeDrift(t *testing.T) {
	t.Parallel()

	w := newWorld(t)
	st := w.applied()
	w.copyFile("files/claude-edited.json", w.home(".claude.json"))

	p, err := Build(w.env, w.targets(), st)
	require.NoError(t, err)

	next, result, err := Pull(w.env, p, st, nil, now)

	require.NoError(t, err)
	require.Len(t, result.Pulled, 1)
	assert.Equal(t, w.home(".claude.json"), result.Pulled[0].Item.Path)
	assert.Equal(t, []string{"mcpServers.notes"}, result.Pulled[0].Paths)
	assert.Empty(t, result.Left)
	assert.Equal(t, golden(t, "claude-pulled.json"), w.read(w.repo("claude/claude.json")))
	assert.Equal(t, w.read(filepath.Join("testdata", "repo", "claude", "claude.json")), w.read(filepath.Join(w.env.Backups(now), w.repo("claude/claude.json"))), "the old source is backed up")
	assert.NotEqual(t, st.Targets[w.home(".claude.json")], next.Targets[w.home(".claude.json")])

	after, err := Build(w.env, w.targets(), next)
	require.NoError(t, err)
	assert.False(t, after.Items[0].Changed(), "the next plan shows the path as unchanged")
}

func TestPullLeavesAConflict(t *testing.T) {
	t.Parallel()

	w := newWorld(t)
	st := w.applied()
	w.copyFile("files/claude-edited.json", w.home(".claude.json"))
	w.copyFile("files/claude-source-changed.json", w.repo("claude/claude.json"))

	p, err := Build(w.env, w.targets(), st)
	require.NoError(t, err)

	next, result, err := Pull(w.env, p, st, nil, now)

	require.NoError(t, err)
	assert.Empty(t, result.Pulled)
	assert.Equal(t, []Left{{Path: w.home(".claude.json"), Reason: "mcpServers.notes changed in both"}}, result.Left)
	assert.Equal(t, st, next)
	assert.Equal(t, w.read(filepath.Join("testdata", "files", "claude-source-changed.json")), w.read(w.repo("claude/claude.json")), "the source is left as it was")
}

func TestPullLeavesALinkTarget(t *testing.T) {
	t.Parallel()

	w := newWorld(t)
	st := w.applied()
	w.replaceLink(w.home(".zshrc"), "files/.zshrc")
	before := w.read(w.repo("zsh/.zshrc"))

	p, err := Build(w.env, w.targets(), st)
	require.NoError(t, err)

	next, result, err := Pull(w.env, p, st, nil, now)

	require.NoError(t, err)
	assert.Empty(t, result.Pulled)
	assert.Equal(t, st, next)
	assert.Equal(t, before, w.read(w.repo("zsh/.zshrc")))
}

func TestPullPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		paths   func(w world) []string
		pulled  int
		left    []string
		wantErr string
	}{
		{
			name:   "only the named target is pulled",
			paths:  func(w world) []string { return []string{w.home(".claude.json")} },
			pulled: 1,
		},
		{
			name:   "a named target with nothing to pull is reported",
			paths:  func(w world) []string { return []string{w.home(".zshrc")} },
			left:   []string{"nothing to pull"},
			pulled: 0,
		},
		{
			name:    "a path that is not a target is an error",
			paths:   func(w world) []string { return []string{w.home(".bashrc")} },
			wantErr: "not a managed target",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w := newWorld(t)
			st := w.applied()
			w.copyFile("files/claude-edited.json", w.home(".claude.json"))

			p, err := Build(w.env, w.targets(), st)
			require.NoError(t, err)

			_, result, err := Pull(w.env, p, st, tc.paths(w), now)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				assert.Equal(t, w.read(filepath.Join("testdata", "repo", "claude", "claude.json")), w.read(w.repo("claude/claude.json")))

				return
			}

			require.NoError(t, err)
			assert.Len(t, result.Pulled, tc.pulled)

			var reasons []string
			for _, l := range result.Left {
				reasons = append(reasons, l.Reason)
			}

			assert.Equal(t, tc.left, reasons)
		})
	}
}

func TestPullStopsOnAnErrorAndKeepsTheState(t *testing.T) {
	t.Parallel()

	w := newWorld(t)
	st := w.applied()
	w.copyFile("files/claude-edited.json", w.home(".claude.json"))

	p, err := Build(w.env, w.targets(), st)
	require.NoError(t, err)

	// A source removed after the plan cannot be written.
	require.NoError(t, os.RemoveAll(w.repo("claude")))

	next, result, err := Pull(w.env, p, st, nil, now)

	require.Error(t, err)
	assert.Empty(t, result.Pulled)
	assert.Equal(t, st, next)
}

func TestRenderPull(t *testing.T) {
	t.Parallel()

	home := "/Users/test"

	tests := []struct {
		name   string
		result PullResult
		golden string
	}{
		{
			name: "pulled and left",
			result: PullResult{
				Pulled: []Pulled{
					{Item: Item{Path: home + "/.claude.json", Strategy: manifest.Merge}, Paths: []string{"mcpServers.notes"}},
				},
				Left: []Left{{Path: home + "/.claude-work/.claude.json", Reason: "mcpServers.notes changed in both"}},
			},
			golden: "pull.txt",
		},
		{name: "nothing", golden: "pull-nothing.txt"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer

			require.NoError(t, RenderPull(&out, tc.result, home))
			assert.Equal(t, golden(t, tc.golden), out.String())
		})
	}
}

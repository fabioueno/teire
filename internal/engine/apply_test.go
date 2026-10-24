package engine

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/fabioueno/teire/internal/manifest"
	"github.com/fabioueno/teire/internal/state"
)

func overwrite(Conflict) (Choice, error) { return Overwrite, nil }
func skip(Conflict) (Choice, error)      { return Skip, nil }

func TestApplyRecordsWhatItWrote(t *testing.T) {
	t.Parallel()

	w := newWorld(t)
	p, err := Build(w.env, w.targets(), emptyState())
	require.NoError(t, err)

	next, result, err := Apply(w.env, p, emptyState(), skip, now)

	require.NoError(t, err)
	assert.Zero(t, result.Skipped)

	dest, err := os.Readlink(w.home(".zshrc"))
	require.NoError(t, err)
	assert.Equal(t, w.repo("zsh/.zshrc"), dest)
	assert.Equal(t, "https://notes.example/mcp", gjson.Get(w.read(w.home(".claude.json")), "mcpServers.notes.url").String())

	assert.Equal(t, state.Entry{Strategy: manifest.Link, Source: w.repo("zsh/.zshrc")}, next.Targets[w.home(".zshrc")])
	assert.Contains(t, next.Targets[w.home(".claude.json")].Owned, "mcpServers.notes")
}

func TestApplyBacksUpAMergeTarget(t *testing.T) {
	t.Parallel()

	w := newWorld(t)
	before := w.read(w.home(".claude.json"))
	p, err := Build(w.env, w.targets(), emptyState())
	require.NoError(t, err)

	_, _, err = Apply(w.env, p, emptyState(), skip, now)

	require.NoError(t, err)
	assert.Equal(t, before, w.read(filepath.Join(w.env.Backups(now), w.home(".claude.json"))))
}

func TestApplyResolvesConflicts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		resolve     Resolver
		wantSkipped int
		wantZshrc   string
		wantURL     string
	}{
		{name: "skip", resolve: skip, wantSkipped: 2, wantZshrc: "export EDITOR=vim\n", wantURL: "https://notes.example/mcp?debug=1"},
		{name: "overwrite", resolve: overwrite, wantZshrc: "export EDITOR=nvim\n", wantURL: "https://notes.example/mcp"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w := newWorld(t)
			w.copyFile("files/.zshrc", w.home(".zshrc"))
			w.copyFile("files/claude-edited.json", w.home(".claude.json"))

			p, err := Build(w.env, w.targets(), emptyState())
			require.NoError(t, err)
			require.Equal(t, 2, p.Conflicts())

			next, result, err := Apply(w.env, p, emptyState(), tc.resolve, now)

			require.NoError(t, err)
			assert.Equal(t, tc.wantSkipped, result.Skipped)
			assert.Equal(t, tc.wantZshrc, w.read(w.home(".zshrc")))
			assert.Equal(t, tc.wantURL, gjson.Get(w.read(w.home(".claude.json")), "mcpServers.notes.url").String())

			_, recorded := next.Targets[w.home(".zshrc")]
			assert.Equal(t, tc.wantSkipped == 0, recorded, "a skipped link is not recorded")
		})
	}
}

func TestApplyStopsOnAResolverError(t *testing.T) {
	t.Parallel()

	w := newWorld(t)
	w.copyFile("files/.zshrc", w.home(".zshrc"))
	p, err := Build(w.env, w.targets(), emptyState())
	require.NoError(t, err)

	failing := func(Conflict) (Choice, error) { return Skip, errors.New("stdin closed") }

	next, _, err := Apply(w.env, p, emptyState(), failing, now)

	assert.ErrorContains(t, err, "stdin closed")
	assert.Contains(t, next.Targets, w.home(".claude.json"), "the merge before the failure is recorded")
}

func TestApplyReportsAMissingMergeTarget(t *testing.T) {
	t.Parallel()

	w := newWorld(t)
	require.NoError(t, os.Remove(w.home(".claude.json")))
	p, err := Build(w.env, w.targets(), emptyState())
	require.NoError(t, err)

	_, result, err := Apply(w.env, p, emptyState(), skip, now)

	require.NoError(t, err)
	assert.Equal(t, []string{w.home(".claude.json")}, result.Missing)
}

func TestApplyForgetsRemovedEntries(t *testing.T) {
	t.Parallel()

	w := newWorld(t)
	require.NoError(t, os.Symlink(w.repo("zsh/.zshrc"), w.home(".zshrc")))
	st := state.State{Targets: map[string]state.Entry{
		w.home(".zshrc"): {Strategy: manifest.Link, Source: w.repo("zsh/.zshrc")},
	}}

	p, err := Build(w.env, nil, st)
	require.NoError(t, err)

	next, _, err := Apply(w.env, p, st, skip, now)

	require.NoError(t, err)
	assert.Empty(t, next.Targets)

	_, err = os.Lstat(w.home(".zshrc"))
	assert.ErrorIs(t, err, os.ErrNotExist)
}

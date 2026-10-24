package engine

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fabioueno/teire/internal/manifest"
	"github.com/fabioueno/teire/internal/state"
	"github.com/fabioueno/teire/internal/strategy"
)

func TestBuildPlansEachTarget(t *testing.T) {
	t.Parallel()

	w := newWorld(t)

	p, err := Build(w.env, w.targets(), emptyState())

	require.NoError(t, err)
	require.Len(t, p.Items, 2)

	merge, link := p.Items[0], p.Items[1]
	assert.Equal(t, w.repo("claude/claude.json"), merge.Source)
	assert.Equal(t, []strategy.Action{strategy.Set}, actions(merge.Merge))
	assert.Equal(t, w.repo("zsh/.zshrc"), link.Source)
	assert.Equal(t, strategy.LinkCreate, link.Link.Action)
	assert.True(t, p.Changed())
	assert.Zero(t, p.Conflicts())
}

func TestBuildMarksAMissingMergeTarget(t *testing.T) {
	t.Parallel()

	w := newWorld(t)
	require.NoError(t, os.Remove(w.home(".claude.json")))

	p, err := Build(w.env, w.targets(), emptyState())

	require.NoError(t, err)
	assert.True(t, p.Items[0].Missing)
	assert.True(t, p.Items[0].Changed())
}

func TestBuildPlansEntriesThatLeftTheManifest(t *testing.T) {
	t.Parallel()

	w := newWorld(t)
	require.NoError(t, os.Symlink(w.repo("zsh/.zshrc"), w.home(".zshrc")))

	st := state.State{Targets: map[string]state.Entry{
		w.home(".zshrc"):       {Strategy: manifest.Link, Source: w.repo("zsh/.zshrc")},
		w.home(".claude.json"): {Strategy: manifest.Merge, Owned: map[string]string{"mcpServers.notes": "4f2d"}},
	}}

	p, err := Build(w.env, nil, st)

	require.NoError(t, err)
	require.Len(t, p.Items, 2)

	merge, link := p.Items[0], p.Items[1]
	assert.True(t, merge.Removed)
	assert.Equal(t, []strategy.Action{strategy.Release}, actions(merge.Merge))
	assert.True(t, link.Removed)
	assert.Equal(t, strategy.LinkRemove, link.Link.Action)
}

func TestBuildRejectsAnUnknownStrategyInState(t *testing.T) {
	t.Parallel()

	w := newWorld(t)
	st := state.State{Targets: map[string]state.Entry{w.home(".old"): {Strategy: "copy"}}}

	_, err := Build(w.env, nil, st)

	assert.ErrorContains(t, err, `unknown strategy "copy" in state`)
}

func TestItemChangedAndConflicts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		item          Item
		wantChanged   bool
		wantConflicts int
	}{
		{
			name: "link in place",
			item: Item{Strategy: manifest.Link, Link: strategy.LinkChange{Action: strategy.LinkUnchanged}},
		},
		{
			name:        "link to create",
			item:        Item{Strategy: manifest.Link, Link: strategy.LinkChange{Action: strategy.LinkCreate}},
			wantChanged: true,
		},
		{
			name:          "folder in conflict",
			item:          Item{Strategy: manifest.Folder, Link: strategy.LinkChange{Action: strategy.LinkConflict}},
			wantChanged:   true,
			wantConflicts: 1,
		},
		{
			name: "merge in place",
			item: Item{Strategy: manifest.Merge, Merge: []strategy.Change{{Action: strategy.Unchanged}}},
		},
		{
			name:          "merge with two conflicts",
			item:          Item{Strategy: manifest.Merge, Merge: []strategy.Change{{Action: strategy.Conflict}, {Action: strategy.Drift}, {Action: strategy.Conflict}}},
			wantChanged:   true,
			wantConflicts: 2,
		},
		{
			name:        "removed entry",
			item:        Item{Strategy: manifest.Link, Removed: true, Link: strategy.LinkChange{Action: strategy.LinkRelease}},
			wantChanged: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.wantChanged, tc.item.Changed())
			assert.Equal(t, tc.wantConflicts, tc.item.Conflicts())
		})
	}
}

func actions(changes []strategy.Change) []strategy.Action {
	out := make([]strategy.Action, len(changes))
	for i, c := range changes {
		out[i] = c.Action
	}

	return out
}

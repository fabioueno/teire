package strategy

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestApplyMerge(t *testing.T) {
	t.Parallel()

	past := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	applied := map[string]string{notes: hashAt(t, "source.json", notes)}

	tests := []struct {
		name   string
		target string
		own    []string
		last   map[string]string
		want   string
		writes bool
	}{
		{name: "set keeps every other byte", target: "target-without.json", own: []string{notes}, want: "applied-set.json", writes: true},
		{name: "delete keeps every other byte", target: "target-same.json", last: applied, want: "target-without.json", writes: true},
		{name: "unchanged writes nothing", target: "target-same.json", own: []string{notes}, last: applied, want: "target-same.json"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), ".claude.json")
			require.NoError(t, os.WriteFile(path, fixture(t, tc.target), 0o600))
			require.NoError(t, os.Chtimes(path, past, past))

			changes, err := PlanMerge(fixture(t, "source.json"), fixture(t, tc.target), tc.own, tc.last)
			require.NoError(t, err)

			notKept, err := ApplyMerge(path, changes)
			require.NoError(t, err)
			assert.Empty(t, notKept)

			got, err := os.ReadFile(path)
			require.NoError(t, err)
			assert.Equal(t, string(fixture(t, tc.want)), string(got))
			assert.True(t, json.Valid(got))

			info, err := os.Stat(path)
			require.NoError(t, err)
			assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
			assert.Equal(t, tc.writes, !info.ModTime().Equal(past))
		})
	}
}

func TestApplyMergeNeedsAnExistingTarget(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), ".claude.json")

	_, err := ApplyMerge(path, []Change{{Path: notes, Action: Set, Source: "{}"}})

	assert.ErrorIs(t, err, ErrTargetMissing)

	_, statErr := os.Stat(path)
	assert.ErrorIs(t, statErr, os.ErrNotExist, "the target is not created")
}

func TestNotKeptReportsValuesAnotherWriterReplaced(t *testing.T) {
	t.Parallel()

	source := gjson.GetBytes(fixture(t, "source.json"), notes).Raw

	got := notKept(fixture(t, "target-edited.json"), []Change{
		{Path: notes, Action: Set, Source: source},
		{Path: "mcpServers.github", Action: Delete},
		{Path: "theme", Action: Unchanged},
	})

	assert.Equal(t, []string{notes, "mcpServers.github"}, got)
}

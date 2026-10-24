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

const notes = "mcpServers.notes"

func fixture(t *testing.T, name string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", "merge", name))
	require.NoError(t, err)

	return data
}

func hashAt(t *testing.T, name, path string) string {
	t.Helper()

	value := gjson.GetBytes(fixture(t, name), path)
	require.True(t, value.Exists(), "%s has %s", name, path)

	return hashValue(value)
}

func TestPlanMerge(t *testing.T) {
	t.Parallel()

	applied := map[string]string{notes: hashAt(t, "source.json", notes)}

	tests := []struct {
		name   string
		source string
		target string
		own    []string
		last   map[string]string
		want   Action
	}{
		{name: "first apply, path absent", source: "source.json", target: "target-without.json", own: []string{notes}, want: Set},
		{name: "first apply, same value", source: "source.json", target: "target-same.json", own: []string{notes}, want: Unchanged},
		{name: "first apply, other value", source: "source.json", target: "target-edited.json", own: []string{notes}, want: Conflict},
		{name: "nothing changed", source: "source.json", target: "target-same.json", own: []string{notes}, last: applied, want: Unchanged},
		{name: "source changed", source: "source-changed.json", target: "target-same.json", own: []string{notes}, last: applied, want: Set},
		{name: "target changed", source: "source.json", target: "target-edited.json", own: []string{notes}, last: applied, want: Drift},
		{name: "target removed the path", source: "source.json", target: "target-without.json", own: []string{notes}, last: applied, want: Drift},
		{name: "both changed apart", source: "source-changed.json", target: "target-edited.json", own: []string{notes}, last: applied, want: Conflict},
		{name: "both changed alike", source: "source-changed.json", target: "target-changed.json", own: []string{notes}, last: applied, want: Unchanged},
		{name: "left own, value as applied", source: "source.json", target: "target-same.json", last: applied, want: Delete},
		{name: "left own, value edited", source: "source.json", target: "target-edited.json", last: applied, want: Release},
		{name: "left own, value gone", source: "source.json", target: "target-without.json", last: applied, want: Release},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			changes, err := PlanMerge(fixture(t, tc.source), fixture(t, tc.target), tc.own, tc.last)

			require.NoError(t, err)
			require.Len(t, changes, 1)
			assert.Equal(t, notes, changes[0].Path)
			assert.Equal(t, tc.want, changes[0].Action)
		})
	}
}

func TestPlanMergeErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		source  string
		target  string
		own     []string
		wantErr string
	}{
		{name: "path not in source", source: "source.json", target: "target-same.json", own: []string{"mcpServers.github"}, wantErr: `owned path "mcpServers.github" is not in the source`},
		{name: "query syntax", source: "source.json", target: "target-same.json", own: []string{"mcpServers.*"}, wantErr: "uses query syntax"},
		{name: "empty path", source: "source.json", target: "target-same.json", own: []string{""}, wantErr: "owned path is empty"},
		{name: "invalid source", source: "invalid.json", target: "target-same.json", own: []string{notes}, wantErr: "source is not valid JSON"},
		{name: "invalid target", source: "source.json", target: "invalid.json", own: []string{notes}, wantErr: "target is not valid JSON"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := PlanMerge(fixture(t, tc.source), fixture(t, tc.target), tc.own, nil)

			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestOwnedRecordsWhatTheNextPlanCompares(t *testing.T) {
	t.Parallel()

	source := gjson.GetBytes(fixture(t, "source.json"), notes).Raw
	last := map[string]string{"drift": "d1", "conflict": "c1", "delete": "x1", "release": "r1"}

	owned := Owned([]Change{
		{Path: "unchanged", Action: Unchanged, Source: source},
		{Path: "set", Action: Set, Source: source},
		{Path: "drift", Action: Drift, Source: source},
		{Path: "conflict", Action: Conflict, Source: source},
		{Path: "new-conflict", Action: Conflict, Source: source},
		{Path: "delete", Action: Delete},
		{Path: "release", Action: Release},
	}, last)

	assert.Equal(t, map[string]string{
		"unchanged": hashRaw(source),
		"set":       hashRaw(source),
		"drift":     "d1",
		"conflict":  "c1",
	}, owned)
}

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

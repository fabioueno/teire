package strategy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

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
		{name: "query syntax in a path", source: "source.json", target: "target-same.json", own: []string{"mcpServers.*"}, wantErr: "uses query syntax"},
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

package strategy

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tailscale/hujson"
)

const (
	jsoncNotes   = "context_servers.notes"
	jsoncDotted  = `projects.a\.b`
	jsoncSpecial = "projects.x/y~z"
)

func TestApplyMergeJSONC(t *testing.T) {
	t.Parallel()

	past := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	applied := map[string]string{jsoncNotes: hashAt(t, "jsonc-source.jsonc", jsoncNotes)}

	tests := []struct {
		name   string
		target string
		own    []string
		last   map[string]string
		want   string
		writes bool
	}{
		{name: "set keeps every other byte", target: "jsonc-target-without.jsonc", own: []string{jsoncNotes, jsoncDotted, jsoncSpecial}, want: "jsonc-applied-set.jsonc", writes: true},
		{name: "delete keeps every other byte", target: "jsonc-target-same.jsonc", last: applied, want: "jsonc-applied-delete.jsonc", writes: true},
		{name: "unchanged writes nothing", target: "jsonc-target-same.jsonc", own: []string{jsoncNotes}, last: applied, want: "jsonc-target-same.jsonc"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "settings.json")
			require.NoError(t, os.WriteFile(path, fixture(t, tc.target), 0o600))
			require.NoError(t, os.Chtimes(path, past, past))

			changes, err := PlanMerge(fixture(t, "jsonc-source.jsonc"), fixture(t, tc.target), tc.own, tc.last)
			require.NoError(t, err)

			merged, err := ApplyMerge(path, changes)
			require.NoError(t, err)
			assert.Empty(t, merged.NotKept)
			assert.Empty(t, merged.Moved)

			got, err := os.ReadFile(path)
			require.NoError(t, err)
			assert.Equal(t, string(fixture(t, tc.want)), string(got))

			_, err = hujson.Parse(got)
			assert.NoError(t, err)

			info, err := os.Stat(path)
			require.NoError(t, err)
			assert.Equal(t, tc.writes, !info.ModTime().Equal(past))
		})
	}
}

func TestApplyMergeJSONCLeavesValuesThatMovedSincePlan(t *testing.T) {
	t.Parallel()

	changes, err := PlanMerge(fixture(t, "jsonc-source.jsonc"), fixture(t, "jsonc-target-without.jsonc"), []string{jsoncNotes}, nil)
	require.NoError(t, err)

	path := filepath.Join(t.TempDir(), "settings.json")
	require.NoError(t, os.WriteFile(path, fixture(t, "jsonc-target-same.jsonc"), 0o600))

	merged, err := ApplyMerge(path, changes)

	require.NoError(t, err)
	assert.Equal(t, []string{jsoncNotes}, merged.Moved)

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, string(fixture(t, "jsonc-target-same.jsonc")), string(got))
}

func TestApplyMergeJSONCNeedsAnExistingTarget(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "settings.json")

	_, err := ApplyMerge(path, []Change{{Path: jsoncNotes, Action: Set, Source: "{}"}})

	assert.ErrorIs(t, err, ErrTargetMissing)
}

func TestHashIgnoresComments(t *testing.T) {
	t.Parallel()

	commented := hashAt(t, "jsonc-source.jsonc", jsoncNotes)

	assert.Equal(t, commented, hashAt(t, "jsonc-target-same.jsonc", jsoncNotes))
	assert.NotEqual(t, commented, hashAt(t, "jsonc-target-same.jsonc", "context_servers.last"))
}

func TestLookup(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "nested key", path: jsoncNotes, want: `{"url": "https://notes.example/mcp", "enabled": true}`},
		{name: "key with a dot", path: jsoncDotted, want: `{ "trusted": true }`},
		{name: "key with a slash and a tilde", path: jsoncSpecial, want: `{ "trusted": true }`},
		{name: "scalar", path: "context_servers.notes.enabled", want: "true"},
		{name: "absent", path: "context_servers.missing"},
		{name: "dotted key without its escape", path: "projects.a.b"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := lookup(fixture(t, "jsonc-source.jsonc"), tc.path)

			assert.Equal(t, tc.want != "", got.Exists())

			if tc.want != "" {
				assert.JSONEq(t, tc.want, got.Raw)
				assert.NotContains(t, got.Raw, "transport", "comments are dropped")
			}
		})
	}
}

func TestSetPathCreatesMissingParents(t *testing.T) {
	t.Parallel()

	got, err := setPath(fixture(t, "jsonc-target-without.jsonc"), "servers.deep.notes", []byte(`{"url":"x"}`))

	require.NoError(t, err)
	assert.JSONEq(t, `{"url":"x"}`, lookup(got, "servers.deep.notes").Raw)
	assert.Equal(t, `"One Dark"`, lookup(got, "theme").Raw)
}

func TestDeletePathOfAbsentPathChangesNothing(t *testing.T) {
	t.Parallel()

	target := fixture(t, "jsonc-target-without.jsonc")

	got, err := deletePath(target, "context_servers.notes")

	require.NoError(t, err)
	assert.Equal(t, string(target), string(got))
}

func TestSplitPathAndPointer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		keys []string
		want string
	}{
		{name: "plain keys", path: "mcpServers.notes", keys: []string{"mcpServers", "notes"}, want: "/mcpServers/notes"},
		{name: "escaped dot", path: `projects.a\.b`, keys: []string{"projects", "a.b"}, want: "/projects/a.b"},
		{name: "slash", path: "projects./Users/test", keys: []string{"projects", "/Users/test"}, want: "/projects/~1Users~1test"},
		{name: "tilde", path: "a~b.c", keys: []string{"a~b", "c"}, want: "/a~0b/c"},
		{name: "tilde before a one", path: "~1", keys: []string{"~1"}, want: "/~01"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			keys := splitPath(tc.path)

			assert.Equal(t, tc.keys, keys)
			assert.Equal(t, tc.want, pointer(keys))
		})
	}
}

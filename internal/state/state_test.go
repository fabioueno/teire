package state

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fabioueno/teire/internal/manifest"
)

var saved = State{Targets: map[string]Entry{
	"/Users/test/.claude.json": {
		Strategy: manifest.Merge,
		Owned:    map[string]string{"mcpServers.notes": "4f2d"},
	},
	"/Users/test/.claude/settings.json": {Strategy: manifest.Link},
}}

func TestLoadWithoutAFileGivesAnEmptyState(t *testing.T) {
	t.Parallel()

	s, err := Load(filepath.Join(t.TempDir(), "state.json"))

	require.NoError(t, err)
	assert.Equal(t, State{Targets: map[string]Entry{}}, s)
}

func TestLoadReadsASavedState(t *testing.T) {
	t.Parallel()

	s, err := Load(filepath.Join("testdata", "saved.json"))

	require.NoError(t, err)
	assert.Equal(t, saved, s)
}

func TestSaveWritesTheFileThatLoadReads(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "teire", "state.json")

	require.NoError(t, saved.Save(path))

	got, err := os.ReadFile(path)
	require.NoError(t, err)

	want, err := os.ReadFile(filepath.Join("testdata", "saved.json"))
	require.NoError(t, err)
	assert.Equal(t, string(want), string(got))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	dir, err := os.Stat(filepath.Dir(path))
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o700), dir.Mode().Perm())
}

func TestLoadErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		file    string
		wantErr string
	}{
		{name: "invalid JSON", file: "invalid.json", wantErr: "parse state"},
		{name: "other version", file: "future-version.json", wantErr: "has version 2, want 1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := Load(filepath.Join("testdata", tc.file))

			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}

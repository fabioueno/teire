package strategy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

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

func linkRepo(t *testing.T) string {
	t.Helper()

	repo, err := filepath.Abs(filepath.Join("testdata", "link", "repo"))
	require.NoError(t, err)

	return repo
}

func linkFixture(t *testing.T, name string) []byte {
	t.Helper()

	data, err := os.ReadFile(filepath.Join("testdata", "link", name))
	require.NoError(t, err)

	return data
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, data, 0o644))
}

func symlink(t *testing.T, source, path string) {
	t.Helper()

	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.Symlink(source, path))
}

const notes = "mcpServers.notes"

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(data)
}

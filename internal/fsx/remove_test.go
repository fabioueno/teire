package fsx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveAllDeletesAFolderAndItsContents(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "20261024T101500Z")
	file := filepath.Join(dir, "home", ".zshrc")
	require.NoError(t, os.MkdirAll(filepath.Dir(file), 0o755))
	require.NoError(t, os.WriteFile(file, []byte("export EDITOR=nvim\n"), 0o644))

	require.NoError(t, RemoveAll(dir))

	_, err := os.Lstat(dir)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestRemoveAllIgnoresAMissingPath(t *testing.T) {
	t.Parallel()

	assert.NoError(t, RemoveAll(filepath.Join(t.TempDir(), "missing")))
}

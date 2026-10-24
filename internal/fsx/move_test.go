package fsx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMoveCreatesTheDestinationFolders(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "home", ".gitconfig")
	dest := filepath.Join(dir, "repo", "git", ".gitconfig")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("[user]\n"), 0o644))

	require.NoError(t, Move(path, dest))

	data, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "[user]\n", string(data))

	_, err = os.Lstat(path)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

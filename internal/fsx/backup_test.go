package fsx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupMovesTheTargetUnderItsOwnPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "home", ".zshrc")
	root := filepath.Join(dir, "backups", "20261024T101500")
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte("export EDITOR=nvim\n"), 0o644))

	dest, err := Backup(path, root)

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(root, path), dest)

	data, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "export EDITOR=nvim\n", string(data))

	_, err = os.Lstat(path)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

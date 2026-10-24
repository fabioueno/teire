package fsx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSymlinkCreatesMissingFolders(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "home", ".config", "zed", "settings.json")

	require.NoError(t, Symlink("/repo/zed/settings.json", path))

	dest, err := os.Readlink(path)
	require.NoError(t, err)
	assert.Equal(t, "/repo/zed/settings.json", dest)
}

func TestRemoveLinkRemovesOnlyLinks(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	link := filepath.Join(dir, "link")
	plain := filepath.Join(dir, "plain")
	require.NoError(t, os.Symlink("/repo/file", link))
	require.NoError(t, os.WriteFile(plain, nil, 0o644))

	require.NoError(t, RemoveLink(link))
	assert.ErrorContains(t, RemoveLink(plain), "not a symlink")

	_, err := os.Lstat(link)
	assert.ErrorIs(t, err, os.ErrNotExist)

	_, err = os.Lstat(plain)
	assert.NoError(t, err)
}

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

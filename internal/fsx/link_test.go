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

package fsx

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteFileAtomicReplacesContentAndSetsMode(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(path, []byte("old"), 0o644))

	require.NoError(t, WriteFileAtomic(path, []byte("new"), 0o600))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "new", string(data))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 1, "no temporary file stays behind")
}

func TestWriteFileAtomicCreatesAMissingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "state.json")

	require.NoError(t, WriteFileAtomic(path, []byte("{}"), 0o600))

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "{}", string(data))
}

func TestWriteFileAtomicWritesThroughASymlink(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	real := filepath.Join(dir, "real.json")
	link := filepath.Join(dir, "link.json")
	require.NoError(t, os.WriteFile(real, []byte("old"), 0o644))
	require.NoError(t, os.Symlink(real, link))

	require.NoError(t, WriteFileAtomic(link, []byte("new"), 0o644))

	info, err := os.Lstat(link)
	require.NoError(t, err)
	assert.Equal(t, os.ModeSymlink, info.Mode().Type(), "the link stays a link")

	data, err := os.ReadFile(real)
	require.NoError(t, err)
	assert.Equal(t, "new", string(data))
}

func TestWriteFileAtomicFailsWithoutTheFolder(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing", "state.json")

	assert.Error(t, WriteFileAtomic(path, []byte("{}"), 0o600))
}

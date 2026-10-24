package strategy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyLinkCreatesTheLink(t *testing.T) {
	t.Parallel()

	repo := linkRepo(t)
	settings := filepath.Join(repo, "claude", "settings.json")
	path := filepath.Join(t.TempDir(), ".claude", "settings.json")

	require.NoError(t, ApplyLink(LinkChange{Path: path, Source: settings, Action: LinkCreate}, t.TempDir()))

	dest, err := os.Readlink(path)
	require.NoError(t, err)
	assert.Equal(t, settings, dest)
}

func TestApplyLinkBacksUpWhatItReplaces(t *testing.T) {
	t.Parallel()

	repo := linkRepo(t)
	settings := filepath.Join(repo, "claude", "settings.json")
	path := filepath.Join(t.TempDir(), ".claude", "settings.json")
	backups := t.TempDir()
	writeFile(t, path, linkFixture(t, "other-settings.json"))

	require.NoError(t, ApplyLink(LinkChange{Path: path, Source: settings, Action: LinkReplace}, backups))

	dest, err := os.Readlink(path)
	require.NoError(t, err)
	assert.Equal(t, settings, dest)

	backedUp, err := os.ReadFile(filepath.Join(backups, path))
	require.NoError(t, err)
	assert.Equal(t, linkFixture(t, "other-settings.json"), backedUp)
}

func TestApplyLinkRemovesTheLink(t *testing.T) {
	t.Parallel()

	repo := linkRepo(t)
	settings := filepath.Join(repo, "claude", "settings.json")
	path := filepath.Join(t.TempDir(), ".claude", "settings.json")
	symlink(t, settings, path)

	require.NoError(t, ApplyLink(LinkChange{Path: path, Source: settings, Action: LinkRemove}, t.TempDir()))

	_, err := os.Lstat(path)
	assert.ErrorIs(t, err, os.ErrNotExist)
}

func TestApplyLinkLeavesConflictsAlone(t *testing.T) {
	t.Parallel()

	repo := linkRepo(t)
	settings := filepath.Join(repo, "claude", "settings.json")
	path := filepath.Join(t.TempDir(), ".claude", "settings.json")
	writeFile(t, path, linkFixture(t, "other-settings.json"))

	for _, action := range []LinkAction{LinkConflict, LinkUnchanged, LinkRelease} {
		require.NoError(t, ApplyLink(LinkChange{Path: path, Source: settings, Action: action}, t.TempDir()))
	}

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, linkFixture(t, "other-settings.json"), data)
}

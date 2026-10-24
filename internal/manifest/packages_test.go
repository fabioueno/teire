package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackagesFillsDefaultsForTheHost(t *testing.T) {
	t.Parallel()

	got, err := Packages(loadManifest(t, "claude.toml"), "work", repoWith("claude"), home)

	require.NoError(t, err)
	assert.Equal(t, []ResolvedPackage{
		{Name: "claude", Source: "claude", Target: home + "/.claude-work", Strategy: Files},
	}, got)
}

func TestPackagesFillsKnownDefaults(t *testing.T) {
	t.Parallel()

	got, err := Packages(loadManifest(t, "resolve/empty.toml"), "", repoWith("nvim", "zsh"), home)

	require.NoError(t, err)
	assert.Equal(t, []ResolvedPackage{
		{Name: "nvim", Source: "nvim", Target: home + "/.config/nvim", Strategy: Folder},
		{Name: "zsh", Source: "zsh", Target: home, Strategy: Files},
	}, got)
}

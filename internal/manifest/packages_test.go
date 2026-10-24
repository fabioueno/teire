package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackagesFillsDefaultsForTheHost(t *testing.T) {
	t.Parallel()

	got, err := Packages(loadManifest(t, "claude.toml"), "work", home)

	require.NoError(t, err)
	assert.Equal(t, []ResolvedPackage{
		{Name: "claude", Source: "claude", Target: home + "/.claude-work", Strategy: Link},
	}, got)
}

func TestPackagesNeedsATarget(t *testing.T) {
	t.Parallel()

	_, err := Packages(loadManifest(t, "invalid/missing-target.toml"), "", home)

	assert.ErrorContains(t, err, `package "claude": target is required`)
}

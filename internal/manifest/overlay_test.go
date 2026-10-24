package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveLeavesTheManifestUnchanged(t *testing.T) {
	t.Parallel()

	m := loadManifest(t, "claude.toml")

	_, err := Resolve(m, "work", repo, home)
	require.NoError(t, err)
	assert.Equal(t, "~/.claude", m.Packages["claude"].Target)
	assert.Equal(t, "~/.claude.json", m.Packages["claude"].Files["claude.json"].Target)
}

func TestWithHost(t *testing.T) {
	t.Parallel()

	m := loadManifest(t, "claude.toml")

	tests := []struct {
		name string
		host string
		want map[string]Package
	}{
		{
			name: "no matching host keeps the base entries",
			host: "MacBook-Pro",
			want: m.Packages,
		},
		{
			name: "host fields replace base fields one by one",
			host: "work",
			want: map[string]Package{
				"claude": {
					Target: "~/.claude-work",
					Files: map[string]File{
						"claude.json": {Strategy: Merge, Target: "~/.claude-work/.claude.json", Own: []string{"mcpServers.notes"}},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, withHost(m, tc.host))
		})
	}
}

func TestWithHostAddsAPackage(t *testing.T) {
	t.Parallel()

	got := withHost(loadManifest(t, "resolve/host-adds-package.toml"), "work")

	assert.Equal(t, map[string]Package{"starship": {Target: "~/.config"}}, got)
}

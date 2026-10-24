package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveLeavesTheManifestUnchanged(t *testing.T) {
	t.Parallel()

	m := loadManifest(t, "claude.toml")

	_, err := Resolve(m, "work", repoWith("claude"), home)
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

			got, err := withHost(m, tc.host, repoWith("claude"))
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestWithHostAddsAPackage(t *testing.T) {
	t.Parallel()

	got, err := withHost(loadManifest(t, "resolve/host-adds-package.toml"), "work", repoWith("myapp"))

	require.NoError(t, err)
	assert.Equal(t, map[string]Package{"myapp": {}}, got)
}

func TestWithHostIgnore(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		host string
		want map[string]Package
	}{
		{
			name: "the base list stops discovery",
			host: "MacBook-Pro",
			want: map[string]Package{"zsh": {}},
		},
		{
			name: "a host list adds to the base list",
			host: "work",
			want: map[string]Package{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := withHost(loadManifest(t, "resolve/ignore-host.toml"), tc.host, repoWith("nvim", "zsh"))

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

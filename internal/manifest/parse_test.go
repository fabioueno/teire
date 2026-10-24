package manifest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseReadsPackagesFilesAndHosts(t *testing.T) {
	t.Parallel()

	m := loadManifest(t, "claude.toml")

	assert.Equal(t, Manifest{
		Packages: map[string]Package{
			"claude": {
				Target: "~/.claude",
				Files: map[string]File{
					"claude.json": {Strategy: Merge, Target: "~/.claude.json", Own: []string{"mcpServers.notes"}},
				},
			},
		},
		Hosts: map[string]Host{
			"work": {Packages: map[string]Package{
				"claude": {
					Target: "~/.claude-work",
					Files: map[string]File{
						"claude.json": {Target: "~/.claude-work/.claude.json"},
					},
				},
			}},
		},
	}, m)
}

func TestParseErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		manifest string
		wantErr  string
	}{
		{name: "unknown key", manifest: "parse/unknown-key.toml", wantErr: "unknown keys: package.claude.tagret"},
		{name: "invalid TOML", manifest: "parse/invalid.toml", wantErr: "parse manifest"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := Parse(openTestdata(t, tc.manifest))

			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}

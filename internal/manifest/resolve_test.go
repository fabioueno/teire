package manifest

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const home = "/Users/test"

var repo = fstest.MapFS{
	"claude/settings.json":   {Data: []byte("{}")},
	"claude/CLAUDE.md":       {Data: []byte("# CLAUDE.md")},
	"claude/claude.json":     {Data: []byte(`{"mcpServers":{}}`)},
	"claude/.DS_Store":       {Data: []byte{0}},
	"zsh/.zshrc":             {Data: []byte("")},
	"zsh/.zsh/10-env.zsh":    {Data: []byte("")},
	"nvim/init.lua":          {Data: []byte("")},
	"nvim/lua/plugins.lua":   {Data: []byte("")},
	"starship/starship.toml": {Data: []byte("")},
}

func TestResolve(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		manifest string
		host     string
		want     []Target
	}{
		{
			name:     "default machine",
			manifest: "claude.toml",
			host:     "MacBook-Pro",
			want: []Target{
				{Package: "claude", Source: "claude/claude.json", Path: home + "/.claude.json", Strategy: Merge, Own: []string{"mcpServers.notes"}},
				{Package: "claude", Source: "claude/CLAUDE.md", Path: home + "/.claude/CLAUDE.md", Strategy: Link},
				{Package: "claude", Source: "claude/settings.json", Path: home + "/.claude/settings.json", Strategy: Link},
			},
		},
		{
			name:     "host overrides package and file targets",
			manifest: "claude.toml",
			host:     "work",
			want: []Target{
				{Package: "claude", Source: "claude/claude.json", Path: home + "/.claude-work/.claude.json", Strategy: Merge, Own: []string{"mcpServers.notes"}},
				{Package: "claude", Source: "claude/CLAUDE.md", Path: home + "/.claude-work/CLAUDE.md", Strategy: Link},
				{Package: "claude", Source: "claude/settings.json", Path: home + "/.claude-work/settings.json", Strategy: Link},
			},
		},
		{
			name:     "host adds a package",
			manifest: "resolve/host-adds-package.toml",
			host:     "work",
			want: []Target{
				{Package: "starship", Source: "starship/starship.toml", Path: home + "/.config/starship.toml", Strategy: Link},
			},
		},
		{
			name:     "nested files keep their folders",
			manifest: "resolve/nested-files.toml",
			want: []Target{
				{Package: "zsh", Source: "zsh/.zsh/10-env.zsh", Path: home + "/.zsh/10-env.zsh", Strategy: Link},
				{Package: "zsh", Source: "zsh/.zshrc", Path: home + "/.zshrc", Strategy: Link},
			},
		},
		{
			name:     "folder strategy gives one target",
			manifest: "resolve/folder.toml",
			want: []Target{
				{Package: "nvim", Source: "nvim", Path: home + "/.config/nvim", Strategy: Folder},
			},
		},
		{
			name:     "source names another folder",
			manifest: "resolve/other-source.toml",
			want: []Target{
				{Package: "prompt", Source: "starship/starship.toml", Path: "/etc/xdg/starship.toml", Strategy: Link},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := Resolve(loadManifest(t, tc.manifest), tc.host, repo, home)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestResolveErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		manifest string
		wantErr  string
	}{
		{
			name:     "missing target",
			manifest: "invalid/missing-target.toml",
			wantErr:  `package "claude": target is required`,
		},
		{
			name:     "relative target",
			manifest: "invalid/relative-target.toml",
			wantErr:  `target ".claude" is not absolute`,
		},
		{
			name:     "missing source folder",
			manifest: "invalid/missing-source.toml",
			wantErr:  `package "git": source folder "git"`,
		},
		{
			name:     "source is a file",
			manifest: "invalid/source-is-file.toml",
			wantErr:  `source "claude/CLAUDE.md" is not a folder`,
		},
		{
			name:     "unknown package strategy",
			manifest: "invalid/unknown-strategy.toml",
			wantErr:  `unknown strategy "copy"`,
		},
		{
			name:     "merge on a package",
			manifest: "invalid/merge-package.toml",
			wantErr:  `strategy "merge" is only valid for a file entry`,
		},
		{
			name:     "file entries on a folder package",
			manifest: "invalid/folder-file-entries.toml",
			wantErr:  `file entries are not valid with strategy "folder"`,
		},
		{
			name:     "file entry without a source file",
			manifest: "invalid/missing-file.toml",
			wantErr:  `file "missing.json": not found in "claude"`,
		},
		{
			name:     "merge without own",
			manifest: "invalid/merge-without-own.toml",
			wantErr:  `file "claude.json": strategy "merge" needs at least one path in own`,
		},
		{
			name:     "own without merge",
			manifest: "invalid/own-without-merge.toml",
			wantErr:  `own is only valid with strategy "merge"`,
		},
		{
			name:     "folder on a file",
			manifest: "invalid/folder-on-file.toml",
			wantErr:  `strategy "folder" is only valid for a package`,
		},
		{
			name:     "two sources for one target",
			manifest: "invalid/duplicate-target.toml",
			wantErr:  `comes from claude/CLAUDE.md (package "claude") and claude/CLAUDE.md (package "copy")`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := Resolve(loadManifest(t, tc.manifest), "", repo, home)
			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestResolveLeavesTheManifestUnchanged(t *testing.T) {
	t.Parallel()

	m := loadManifest(t, "claude.toml")

	_, err := Resolve(m, "work", repo, home)
	require.NoError(t, err)
	assert.Equal(t, "~/.claude", m.Packages["claude"].Target)
	assert.Equal(t, "~/.claude.json", m.Packages["claude"].Files["claude.json"].Target)
}

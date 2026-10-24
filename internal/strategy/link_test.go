package strategy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanLink(t *testing.T) {
	t.Parallel()

	repo := linkRepo(t)
	settings := filepath.Join(repo, "claude", "settings.json")
	nvim := filepath.Join(repo, "nvim")

	tests := []struct {
		name       string
		target     string
		source     string
		folder     bool
		setup      func(t *testing.T, home string)
		want       LinkAction
		wantReason string
	}{
		{
			name:   "missing file",
			target: ".claude/settings.json",
			source: settings,
			want:   LinkCreate,
		},
		{
			name:   "link already in place",
			target: ".claude/settings.json",
			source: settings,
			setup: func(t *testing.T, home string) {
				symlink(t, settings, filepath.Join(home, ".claude", "settings.json"))
			},
			want: LinkUnchanged,
		},
		{
			name:   "link to the same file through another path",
			target: ".claude/settings.json",
			source: settings,
			setup: func(t *testing.T, home string) {
				symlink(t, repo, filepath.Join(home, ".dotfiles"))
				symlink(t, filepath.Join(home, ".dotfiles", "claude", "settings.json"), filepath.Join(home, ".claude", "settings.json"))
			},
			want:       LinkReplace,
			wantReason: "links to the same file through",
		},
		{
			name:   "link to another file",
			target: ".claude/settings.json",
			source: settings,
			setup: func(t *testing.T, home string) {
				other := filepath.Join(home, "other-settings.json")
				writeFile(t, other, linkFixture(t, "other-settings.json"))
				symlink(t, other, filepath.Join(home, ".claude", "settings.json"))
			},
			want:       LinkConflict,
			wantReason: "links to ",
		},
		{
			name:   "link to a path that no longer exists",
			target: ".claude/settings.json",
			source: settings,
			setup: func(t *testing.T, home string) {
				symlink(t, filepath.Join(home, ".dotfiles", "settings.json"), filepath.Join(home, ".claude", "settings.json"))
			},
			want:       LinkReplace,
			wantReason: "links to a missing path",
		},
		{
			name:   "plain file with the same content",
			target: ".claude/settings.json",
			source: settings,
			setup: func(t *testing.T, home string) {
				writeFile(t, filepath.Join(home, ".claude", "settings.json"), linkFixture(t, "repo/claude/settings.json"))
			},
			want:       LinkReplace,
			wantReason: "has the same content",
		},
		{
			name:   "plain file with other content",
			target: ".claude/settings.json",
			source: settings,
			setup: func(t *testing.T, home string) {
				writeFile(t, filepath.Join(home, ".claude", "settings.json"), linkFixture(t, "other-settings.json"))
			},
			want:       LinkConflict,
			wantReason: "has different content",
		},
		{
			name:   "folder where a file goes",
			target: ".claude/settings.json",
			source: settings,
			setup: func(t *testing.T, home string) {
				require.NoError(t, os.MkdirAll(filepath.Join(home, ".claude", "settings.json"), 0o755))
			},
			want:       LinkConflict,
			wantReason: "is a folder",
		},
		{
			name:   "parent folder links into the repository",
			target: ".claude/settings.json",
			source: settings,
			setup: func(t *testing.T, home string) {
				symlink(t, filepath.Join(repo, "claude"), filepath.Join(home, ".claude"))
			},
			want:       LinkConflict,
			wantReason: "a parent folder links into the repository",
		},
		{
			name:   "missing file under a parent that links into the repository",
			target: ".claude/agents/reviewer.md",
			source: settings,
			setup: func(t *testing.T, home string) {
				symlink(t, filepath.Join(repo, "claude"), filepath.Join(home, ".claude"))
			},
			want:       LinkConflict,
			wantReason: "a parent folder links into the repository",
		},
		{
			name:   "missing folder",
			target: ".config/nvim",
			source: nvim,
			folder: true,
			want:   LinkCreate,
		},
		{
			name:   "folder link already in place",
			target: ".config/nvim",
			source: nvim,
			folder: true,
			setup: func(t *testing.T, home string) {
				symlink(t, nvim, filepath.Join(home, ".config", "nvim"))
			},
			want: LinkUnchanged,
		},
		{
			name:   "folder link through another path",
			target: ".config/nvim",
			source: nvim,
			folder: true,
			setup: func(t *testing.T, home string) {
				symlink(t, repo, filepath.Join(home, ".dotfiles"))
				symlink(t, filepath.Join(home, ".dotfiles", "nvim"), filepath.Join(home, ".config", "nvim"))
			},
			want:       LinkReplace,
			wantReason: "links to the same file through",
		},
		{
			name:   "real folder where a folder link goes",
			target: ".config/nvim",
			source: nvim,
			folder: true,
			setup: func(t *testing.T, home string) {
				writeFile(t, filepath.Join(home, ".config", "nvim", "init.lua"), linkFixture(t, "repo/nvim/init.lua"))
			},
			want:       LinkConflict,
			wantReason: "is a folder with its own files",
		},
		{
			name:   "file where a folder link goes",
			target: ".config/nvim",
			source: nvim,
			folder: true,
			setup: func(t *testing.T, home string) {
				writeFile(t, filepath.Join(home, ".config", "nvim"), linkFixture(t, "repo/nvim/init.lua"))
			},
			want:       LinkConflict,
			wantReason: "is a file",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			home := t.TempDir()
			if tc.setup != nil {
				tc.setup(t, home)
			}

			path := filepath.Join(home, tc.target)

			c, err := PlanLink(path, tc.source, repo, tc.folder)

			require.NoError(t, err)
			assert.Equal(t, path, c.Path)
			assert.Equal(t, tc.source, c.Source)
			assert.Equal(t, tc.want, c.Action)
			assert.Contains(t, c.Reason, tc.wantReason)
		})
	}
}

func TestPlanUnlink(t *testing.T) {
	t.Parallel()

	repo := linkRepo(t)
	settings := filepath.Join(repo, "claude", "settings.json")

	tests := []struct {
		name  string
		setup func(t *testing.T, path string)
		want  LinkAction
	}{
		{
			name:  "link that teire made",
			setup: func(t *testing.T, path string) { symlink(t, settings, path) },
			want:  LinkRemove,
		},
		{
			name: "link changed since",
			setup: func(t *testing.T, path string) {
				symlink(t, filepath.Join(repo, "nvim", "init.lua"), path)
			},
			want: LinkRelease,
		},
		{
			name: "plain file in its place",
			setup: func(t *testing.T, path string) {
				writeFile(t, path, linkFixture(t, "other-settings.json"))
			},
			want: LinkRelease,
		},
		{
			name:  "nothing there",
			setup: func(*testing.T, string) {},
			want:  LinkRelease,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), ".claude", "settings.json")
			tc.setup(t, path)

			assert.Equal(t, tc.want, PlanUnlink(path, settings).Action)
		})
	}
}

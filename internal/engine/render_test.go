package engine

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fabioueno/teire/internal/manifest"
	"github.com/fabioueno/teire/internal/strategy"
)

const renderHome = "/Users/test"

func TestShortPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		want string
	}{
		{path: "/Users/test", want: "~"},
		{path: "/Users/test/.zshrc", want: "~/.zshrc"},
		{path: "/Users/test/.config/zed/settings.json", want: "~/.config/zed/settings.json"},
		{path: "/Users/tester/.zshrc", want: "/Users/tester/.zshrc"},
		{path: "/etc/xdg/starship.toml", want: "/etc/xdg/starship.toml"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, ShortPath(tc.path, renderHome))
		})
	}
}

func TestRenderGroupsByPackage(t *testing.T) {
	t.Parallel()

	p := Plan{Items: []Item{
		{
			Package:  "claude",
			Path:     renderHome + "/.claude.json",
			Strategy: manifest.Merge,
			Merge: []strategy.Change{
				{Path: "mcpServers.notes", Action: strategy.Set},
				{Path: "theme", Action: strategy.Unchanged},
			},
		},
		{
			Package:  "claude",
			Path:     renderHome + "/.claude/settings.json",
			Strategy: manifest.Files,
			Link:     strategy.LinkChange{Action: strategy.LinkUnchanged},
		},
		{
			Package:  "zsh",
			Path:     renderHome + "/.zshrc",
			Strategy: manifest.Files,
			Link:     strategy.LinkChange{Action: strategy.LinkConflict, Reason: strategy.ReasonDifferentContent},
		},
		{
			Package:  "zsh",
			Path:     renderHome + "/.zprofile",
			Strategy: manifest.Merge,
			Missing:  true,
		},
		{
			Path:     renderHome + "/.config/nvim",
			Strategy: manifest.Folder,
			Removed:  true,
			Link:     strategy.LinkChange{Action: strategy.LinkRemove},
		},
	}}

	var out bytes.Buffer
	require.NoError(t, Render(&out, p, renderHome))

	assert.Equal(t, golden(t, "render.txt"), out.String())
}

func TestRenderWhenInPlace(t *testing.T) {
	t.Parallel()

	p := Plan{Items: []Item{
		{Package: "zsh", Path: renderHome + "/.zshrc", Strategy: manifest.Files, Link: strategy.LinkChange{Action: strategy.LinkUnchanged}},
	}}

	var out bytes.Buffer
	require.NoError(t, Render(&out, p, renderHome))

	assert.Equal(t, "Everything is in place (1 targets).\n", out.String())
}

func TestRenderShowsEachPackageOnce(t *testing.T) {
	t.Parallel()

	create := strategy.LinkChange{Action: strategy.LinkCreate}
	p := Plan{Items: []Item{
		{Package: "zsh", Path: renderHome + "/.hushlogin", Strategy: manifest.Files, Link: create},
		{Package: "git", Path: renderHome + "/.gitconfig", Strategy: manifest.Files, Link: create},
		{Package: "zsh", Path: renderHome + "/.zshrc", Strategy: manifest.Files, Link: create},
	}}

	var out bytes.Buffer
	require.NoError(t, Render(&out, p, renderHome))

	assert.Equal(t, golden(t, "render-grouped.txt"), out.String())
}

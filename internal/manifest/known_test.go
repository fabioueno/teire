package manifest

import (
	"maps"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultsFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want Package
	}{
		{name: "zsh", want: Package{Target: "~", Strategy: Files}},
		{name: "nvim", want: Package{Target: "~/.config/nvim", Strategy: Folder}},
		{name: "starship", want: Package{Target: "~/.config", Strategy: Files}},
		{name: "tmux", want: Package{Target: "~/.config/tmux", Strategy: Files}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, defaultsFor(tc.name))
		})
	}
}

func TestDiscover(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		declared map[string]Package
		folders  []string
		want     map[string]Package
	}{
		{
			name:    "a known folder becomes a package",
			folders: []string{"zsh"},
			want:    map[string]Package{"zsh": {}},
		},
		{
			name:     "a declared package keeps its entry",
			declared: map[string]Package{"zsh": {Target: "/etc/zsh"}},
			folders:  []string{"zsh"},
			want:     map[string]Package{"zsh": {Target: "/etc/zsh"}},
		},
		{
			name:    "an unknown folder is left alone",
			folders: []string{"tmux"},
			want:    map[string]Package{},
		},
		{
			name:     "a folder another package uses as its source is left alone",
			declared: map[string]Package{"prompt": {Source: "starship", Target: "/etc/xdg"}},
			folders:  []string{"starship"},
			want:     map[string]Package{"prompt": {Source: "starship", Target: "/etc/xdg"}},
		},
		{
			name: "a known name without a folder is left alone",
			want: map[string]Package{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			packages := map[string]Package{}
			maps.Copy(packages, tc.declared)

			discover(packages, repoWith(tc.folders...))

			assert.Equal(t, tc.want, packages)
		})
	}
}

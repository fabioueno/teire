package engine

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fabioueno/teire/internal/manifest"
)

func TestCoveringPackage(t *testing.T) {
	t.Parallel()

	packages := []manifest.ResolvedPackage{
		{Name: "zsh", Source: "zsh", Target: "/Users/test", Strategy: manifest.Link},
		{Name: "zed", Source: "zed", Target: "/Users/test/.config/zed", Strategy: manifest.Link},
		{Name: "nvim", Source: "nvim", Target: "/Users/test/.config/nvim", Strategy: manifest.Folder},
	}

	tests := []struct {
		name    string
		path    string
		want    string
		wantRel string
		found   bool
	}{
		{name: "the deepest target wins", path: "/Users/test/.config/zed/keymap.json", want: "zed", wantRel: "keymap.json", found: true},
		{name: "home covers a dotfile", path: "/Users/test/.zprofile", want: "zsh", wantRel: ".zprofile", found: true},
		{name: "a folder package is skipped", path: "/Users/test/.config/nvim/init.lua", want: "zsh", wantRel: ".config/nvim/init.lua", found: true},
		{name: "outside every target", path: "/etc/hosts"},
		{name: "a target itself is not covered", path: "/Users/test"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			pkg, rel, found := coveringPackage(packages, tc.path)

			assert.Equal(t, tc.found, found)
			assert.Equal(t, tc.want, pkg.Name)
			assert.Equal(t, tc.wantRel, rel)
		})
	}
}

func TestAdopt(t *testing.T) {
	t.Parallel()

	w := newWorld(t)
	path := w.home(".zprofile")
	w.copyFile("files/.zprofile", path)
	packages := []manifest.ResolvedPackage{{Name: "zsh", Source: "zsh", Target: w.env.Home, Strategy: manifest.Link}}
	st := emptyState()

	dest, err := Adopt(w.env, packages, st, path)

	require.NoError(t, err)
	assert.Equal(t, w.repo("zsh/.zprofile"), dest)

	link, err := os.Readlink(path)
	require.NoError(t, err)
	assert.Equal(t, dest, link)
	assert.Equal(t, dest, st.Targets[path].Source)
}

func TestAdoptErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(w world) string
		wantErr string
	}{
		{
			name:    "missing file",
			setup:   func(w world) string { return w.home(".zprofile") },
			wantErr: "no such file",
		},
		{
			name: "a link",
			setup: func(w world) string {
				require.NoError(t, os.Symlink(w.repo("zsh/.zshrc"), w.home(".zprofile")))
				return w.home(".zprofile")
			},
			wantErr: "only a plain file can be adopted",
		},
		{
			name: "the repository already has the file",
			setup: func(w world) string {
				w.copyFile("files/.zshrc", w.home(".zshrc"))
				return w.home(".zshrc")
			},
			wantErr: "already exists",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			w := newWorld(t)
			packages := []manifest.ResolvedPackage{{Name: "zsh", Source: "zsh", Target: w.env.Home, Strategy: manifest.Link}}

			_, err := Adopt(w.env, packages, emptyState(), tc.setup(w))

			assert.ErrorContains(t, err, tc.wantErr)
		})
	}
}

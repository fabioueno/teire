package cli

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdoptMovesTheFileAndLinksIt(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	path := filepath.Join(h.home, ".zprofile")
	h.copyFile("files/.zprofile", path)

	out, err := h.run("adopt", path)

	require.NoError(t, err)
	assert.Equal(t, "Adopted ~/.zprofile into zsh/.zprofile.\n", out)

	dest := filepath.Join(h.repo, "zsh", ".zprofile")
	assert.Equal(t, dest, h.readlink(path))
	assert.Equal(t, "export PATH=\"$HOME/bin:$PATH\"\n", h.read(dest))

	out, err = h.run("plan")
	assert.Equal(t, 1, exitCode(t, err))
	assert.NotContains(t, out, ".zprofile", "the adopted file is in sync")
}

func TestAdoptNeedsACoveringPackage(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	path := filepath.Join(h.root, "elsewhere", "notes.txt")
	h.copyFile("files/.zprofile", path)

	_, err := h.run("adopt", path)

	assert.ErrorContains(t, err, "no files package has a target that covers it")
}

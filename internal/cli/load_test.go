package cli

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRepositoryIsFoundUpward(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.wd = filepath.Join(h.repo, "claude")

	out, err := h.run("plan")

	assert.Equal(t, 1, exitCode(t, err))
	assert.Equal(t, golden(t, "plan-fresh.txt"), out)
}

func TestRepositoryIsNeeded(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.wd = h.home

	_, err := h.run("plan")

	assert.ErrorContains(t, err, "no teire.toml in this folder or above")
}

func TestRepositoryFromTheFlag(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.wd = h.home

	out, err := h.run("plan", "--repo", h.repo)
	assert.Equal(t, 1, exitCode(t, err))
	assert.Equal(t, golden(t, "plan-fresh.txt"), out)

	_, err = h.run("plan", "--repo", h.home)
	assert.ErrorContains(t, err, "repository "+h.home)
}

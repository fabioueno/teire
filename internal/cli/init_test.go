package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitSetsTheHost(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	out, err := h.run("init", "--host", "work")
	require.NoError(t, err)
	assert.Contains(t, out, `This machine is "work"`)
	assert.Equal(t, "work\n", h.read(h.env().HostFile()))

	out, err = h.run("plan")
	assert.Equal(t, 1, exitCode(t, err))
	assert.Contains(t, out, "  merge     ~/.claude-work/.claude.json\n")
	assert.Contains(t, out, "  create    ~/.claude-work/settings.json\n")
}

func TestInitNeedsAHost(t *testing.T) {
	t.Parallel()

	_, err := newHarness(t).run("init")

	assert.ErrorContains(t, err, "init needs --host")
}

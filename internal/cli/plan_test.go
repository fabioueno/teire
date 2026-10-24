package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlanOnAFreshHome(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	out, err := h.run("plan")

	assert.Equal(t, 1, exitCode(t, err))
	_, show := ExitCode(err)
	assert.False(t, show, "plan's exit code needs no message")
	assert.Equal(t, golden(t, "plan-fresh.txt"), out)
}

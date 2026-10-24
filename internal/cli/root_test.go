package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlanIsTheDefaultCommand(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	out, err := h.run()

	assert.Equal(t, 1, exitCode(t, err))
	assert.Equal(t, golden(t, "plan-fresh.txt"), out)
}

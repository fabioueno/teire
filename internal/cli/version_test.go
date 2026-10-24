package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionPrintsTheBuildVersion(t *testing.T) {
	root := NewRoot()
	var out bytes.Buffer

	root.SetOut(&out)
	root.SetArgs([]string{"version"})

	require.NoError(t, root.Execute())
	assert.Equal(t, "dev\n", out.String())
}

func TestVersionRejectsArguments(t *testing.T) {
	root := NewRoot()
	root.SetArgs([]string{"version", "extra"})

	assert.Error(t, root.Execute())
}

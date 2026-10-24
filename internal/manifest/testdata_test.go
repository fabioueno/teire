package manifest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func openTestdata(t *testing.T, name string) *os.File {
	t.Helper()

	f, err := os.Open(filepath.Join("testdata", name))
	require.NoError(t, err)

	t.Cleanup(func() { _ = f.Close() })

	return f
}

func loadManifest(t *testing.T, name string) Manifest {
	t.Helper()

	m, err := Parse(openTestdata(t, name))
	require.NoError(t, err)

	return m
}

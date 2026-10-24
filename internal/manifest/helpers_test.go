package manifest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

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

// repoWith gives the part of repo below the named top-level folders, so a
// test sees only the known folders it is about.
func repoWith(folders ...string) fstest.MapFS {
	keep := make(map[string]bool, len(folders))
	for _, f := range folders {
		keep[f] = true
	}

	out := fstest.MapFS{}
	for name, file := range repo {
		if keep[strings.SplitN(name, "/", 2)[0]] {
			out[name] = file
		}
	}

	return out
}

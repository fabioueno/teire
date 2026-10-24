package strategy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullMergeWritesTheTargetValueAndKeepsTheRest(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "claude.json")
	require.NoError(t, os.WriteFile(path, fixture(t, "source.json"), 0o640))

	last := map[string]string{notes: hashAt(t, "source.json", notes)}
	changes, err := PlanMerge(fixture(t, "source.json"), fixture(t, "target-changed.json"), []string{notes}, last)
	require.NoError(t, err)
	require.Equal(t, Drift, changes[0].Action)

	require.NoError(t, PullMerge(path, changes))

	assert.Equal(t, string(fixture(t, "pulled-source.json")), readFile(t, path))

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o640), info.Mode().Perm())

	again, err := PlanMerge([]byte(readFile(t, path)), fixture(t, "target-changed.json"), []string{notes}, nil)
	require.NoError(t, err)
	assert.Equal(t, Unchanged, again[0].Action, "the source now holds the target value")
}

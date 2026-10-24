package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// seedBackups makes one backup folder, with a file inside, for each name.
func (h *harness) seedBackups(names ...string) {
	h.t.Helper()

	for _, name := range names {
		dir := filepath.Join(h.env().BackupsDir(), name, "home")
		require.NoError(h.t, os.MkdirAll(dir, 0o755))
		require.NoError(h.t, os.WriteFile(filepath.Join(dir, ".zshrc"), []byte("old\n"), 0o644))
	}
}

func (h *harness) backupNames() []string {
	h.t.Helper()

	entries, err := os.ReadDir(h.env().BackupsDir())
	require.NoError(h.t, err)

	names := []string{}

	for _, e := range entries {
		names = append(names, e.Name())
	}

	return names
}

func TestPrune(t *testing.T) {
	t.Parallel()

	seeded := []string{"20260101T000000Z", "20260102T000000Z", "20260103T000000Z", "20260104T000000Z"}

	tests := []struct {
		name    string
		args    []string
		removed []string
		summary string
	}{
		{
			name:    "keeps ten by default",
			args:    []string{"prune"},
			summary: "Nothing to prune (4 backups, keeping 10).\n",
		},
		{
			name:    "removes all but the newest",
			args:    []string{"prune", "--keep", "2"},
			removed: seeded[:2],
			summary: "Removed 2 backups, kept 2.\n",
		},
		{
			name:    "keep zero removes them all",
			args:    []string{"prune", "--keep", "0"},
			removed: seeded,
			summary: "Removed 4 backups, kept 0.\n",
		},
		{
			name:    "removes a single backup",
			args:    []string{"prune", "--keep", "3"},
			removed: seeded[:1],
			summary: "Removed 1 backup, kept 3.\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := newHarness(t)
			h.seedBackups(seeded...)

			want := ""
			for _, name := range tt.removed {
				want += "  removed   " + filepath.Join(h.env().BackupsDir(), name) + "\n"
			}

			out, err := h.run(tt.args...)

			require.NoError(t, err)
			assert.Equal(t, want+tt.summary, out)
			assert.Equal(t, seeded[len(tt.removed):], h.backupNames())
		})
	}
}

func TestPruneLeavesOtherEntriesAlone(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.seedBackups("20260101T000000Z", "20260102T000000Z", "notes", "2026-01-01")
	require.NoError(t, os.WriteFile(filepath.Join(h.env().BackupsDir(), "20260103T000000Z"), nil, 0o644))

	_, err := h.run("prune", "--keep", "0")

	require.NoError(t, err)
	assert.Equal(t, []string{"2026-01-01", "20260103T000000Z", "notes"}, h.backupNames())
}

func TestPruneWorksWithoutABackupsFolder(t *testing.T) {
	t.Parallel()

	out, err := newHarness(t).run("prune")

	require.NoError(t, err)
	assert.Equal(t, "Nothing to prune (0 backups, keeping 10).\n", out)
}

func TestPruneNeedsNoRepository(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.wd = h.root
	h.seedBackups("20260101T000000Z")

	out, err := h.run("prune", "--keep", "0")

	require.NoError(t, err)
	assert.Contains(t, out, "Removed 1 backup, kept 0.")
}

func TestPruneRejectsANegativeKeep(t *testing.T) {
	t.Parallel()

	_, err := newHarness(t).run("prune", "--keep", "-1")

	assert.ErrorContains(t, err, "--keep must not be negative")
}

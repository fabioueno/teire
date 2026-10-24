package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPullMergeDrift(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	_, err := h.run("apply", "--yes")
	require.NoError(t, err)

	h.copyFile("files/claude-edited.json", filepath.Join(h.home, ".claude.json"))

	out, err := h.run("pull")

	require.NoError(t, err)
	assert.Equal(t, "  pull      ~/.claude.json  (mcpServers.notes)\n1 pulled, 0 left.\n", out)
	assert.Equal(t, golden(t, "claude-pulled.json"), h.read(filepath.Join(h.repo, "claude", "claude.json")))

	_, err = h.run("plan")
	assert.Equal(t, 0, exitCode(t, err), "the next plan shows the path as unchanged")
}

func TestPullLeavesAReplacedLink(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	_, err := h.run("apply", "--yes")
	require.NoError(t, err)

	h.replaceLink(filepath.Join(h.home, ".zshrc"), "files/.zshrc")

	out, err := h.run("pull")

	require.NoError(t, err)
	assert.Equal(t, "Nothing to pull.\n", out)
	assert.Equal(t, "export EDITOR=nvim\n", h.read(filepath.Join(h.repo, "zsh", ".zshrc")))

	_, err = h.run("plan")
	assert.Equal(t, 2, exitCode(t, err), "it is still a conflict")
}

func TestPullLeavesAConflict(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	_, err := h.run("apply", "--yes")
	require.NoError(t, err)

	h.copyFile("files/claude-edited.json", filepath.Join(h.home, ".claude.json"))
	h.copyFile("files/claude-source-changed.json", filepath.Join(h.repo, "claude", "claude.json"))

	out, err := h.run("pull")

	require.NoError(t, err)
	assert.Equal(t, "  left      ~/.claude.json  (mcpServers.notes changed in both)\n0 pulled, 1 left.\n", out)
	assert.Equal(t, h.read(filepath.Join("testdata", "files", "claude-source-changed.json")), h.read(filepath.Join(h.repo, "claude", "claude.json")))
}

func TestPullWithPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    func(h *harness) []string
		want    string
		wantErr string
	}{
		{
			name: "an absolute path",
			args: func(h *harness) []string { return []string{filepath.Join(h.home, ".claude.json")} },
			want: "  pull      ~/.claude.json  (mcpServers.notes)\n1 pulled, 0 left.\n",
		},
		{
			name: "a relative path",
			args: func(h *harness) []string {
				rel, err := filepath.Rel(mustGetwd(t), filepath.Join(h.home, ".claude.json"))
				require.NoError(t, err)

				return []string{rel}
			},
			want: "  pull      ~/.claude.json  (mcpServers.notes)\n1 pulled, 0 left.\n",
		},
		{
			name: "a target with nothing to pull",
			args: func(h *harness) []string { return []string{filepath.Join(h.home, ".zshrc")} },
			want: "  left      ~/.zshrc  (nothing to pull)\n0 pulled, 1 left.\n",
		},
		{
			name:    "a path that is not a target",
			args:    func(h *harness) []string { return []string{filepath.Join(h.home, ".bashrc")} },
			wantErr: "not a managed target",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newHarness(t)
			_, err := h.run("apply", "--yes")
			require.NoError(t, err)

			h.copyFile("files/claude-edited.json", filepath.Join(h.home, ".claude.json"))

			out, err := h.run(append([]string{"pull"}, tc.args(h)...)...)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, out)
		})
	}
}

func mustGetwd(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)

	return wd
}

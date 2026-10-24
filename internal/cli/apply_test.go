package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestApplyThenPlanIsInSync(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	out, err := h.run("apply", "--yes")
	require.NoError(t, err)
	assert.Contains(t, out, "Applied.\n")

	assert.Equal(t, filepath.Join(h.repo, "zsh", ".zshrc"), h.readlink(filepath.Join(h.home, ".zshrc")))
	assert.Equal(t, filepath.Join(h.repo, "nvim"), h.readlink(filepath.Join(h.home, ".config", "nvim")))
	assert.Equal(t, filepath.Join(h.repo, "claude", "settings.json"), h.readlink(filepath.Join(h.home, ".claude", "settings.json")))

	claudeJSON := h.read(filepath.Join(h.home, ".claude.json"))
	assert.Equal(t, "https://notes.example/mcp", gjson.Get(claudeJSON, "mcpServers.notes.url").String())
	assert.Equal(t, int64(3), gjson.Get(claudeJSON, "numStartups").Int())

	out, err = h.run("plan")
	require.NoError(t, err)
	assert.Equal(t, golden(t, "plan-in-sync.txt"), out)
}

func TestApplyWithoutYesNeedsATerminal(t *testing.T) {
	t.Parallel()

	h := newHarness(t)

	_, err := h.run("apply")

	assert.ErrorContains(t, err, "pass --yes")

	_, statErr := os.Lstat(filepath.Join(h.home, ".zshrc"))
	assert.ErrorIs(t, statErr, os.ErrNotExist)
}

func TestApplyAsksOnATerminal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		answer  string
		applied bool
	}{
		{name: "no", answer: "n\n", applied: false},
		{name: "empty answer", answer: "\n", applied: false},
		{name: "yes", answer: "y\n", applied: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newHarness(t)
			h.terminal = true
			h.stdin = tc.answer

			out, err := h.run("apply")

			require.NoError(t, err)
			assert.Contains(t, out, "Apply these changes? [y/N]")

			_, statErr := os.Lstat(filepath.Join(h.home, ".zshrc"))
			assert.Equal(t, tc.applied, statErr == nil)
		})
	}
}

func TestConflictWithoutATerminalIsLeftAlone(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.copyFile("files/.zshrc", filepath.Join(h.home, ".zshrc"))

	out, err := h.run("plan")
	assert.Equal(t, 2, exitCode(t, err))
	assert.Contains(t, out, "  conflict  ~/.zshrc  (has different content)\n")

	_, err = h.run("apply", "--yes")
	assert.Equal(t, 2, exitCode(t, err))
	assert.ErrorContains(t, err, "1 conflicts left")

	assert.Equal(t, "export EDITOR=vim\n", h.read(filepath.Join(h.home, ".zshrc")))
	assert.Equal(t, filepath.Join(h.repo, "nvim"), h.readlink(filepath.Join(h.home, ".config", "nvim")))
}

func TestConflictOverwriteOnATerminal(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	h.copyFile("files/.zshrc", filepath.Join(h.home, ".zshrc"))
	h.terminal = true
	h.stdin = "d\no\n"

	out, err := h.run("apply", "--yes")

	require.NoError(t, err)
	assert.Contains(t, out, "-export EDITOR=vim\n+export EDITOR=nvim\n")
	assert.Equal(t, filepath.Join(h.repo, "zsh", ".zshrc"), h.readlink(filepath.Join(h.home, ".zshrc")))

	backup := filepath.Join(h.env().Backups(now), h.home, ".zshrc")
	assert.Equal(t, "export EDITOR=vim\n", h.read(backup))
}

func TestMissingMergeTargetIsReported(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	require.NoError(t, os.Remove(filepath.Join(h.home, ".claude.json")))

	out, err := h.run("plan")
	assert.Equal(t, 1, exitCode(t, err))
	assert.Contains(t, out, "  missing   ~/.claude.json  (the target does not exist; start its program once)\n")

	out, err = h.run("apply", "--yes")
	require.NoError(t, err)
	assert.Contains(t, out, "~/.claude.json does not exist yet")

	_, statErr := os.Stat(filepath.Join(h.home, ".claude.json"))
	assert.ErrorIs(t, statErr, os.ErrNotExist, "apply does not create a merge target")
}

func TestPackageWhoseFolderIsGoneIsUnlinked(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	_, err := h.run("apply", "--yes")
	require.NoError(t, err)

	require.NoError(t, os.RemoveAll(filepath.Join(h.repo, "nvim")))

	out, err := h.run("plan")
	assert.Equal(t, 1, exitCode(t, err))
	assert.Contains(t, out, "(no longer managed)\n  remove    ~/.config/nvim\n")

	_, err = h.run("apply", "--yes")
	require.NoError(t, err)

	_, statErr := os.Lstat(filepath.Join(h.home, ".config", "nvim"))
	assert.ErrorIs(t, statErr, os.ErrNotExist)

	_, err = h.run("plan")
	assert.NoError(t, err)
}

func TestMergeConflict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		terminal bool
		stdin    string
		wantURL  string
		wantCode int
	}{
		{name: "left alone without a terminal", wantURL: "https://notes.example/mcp?debug=1", wantCode: 2},
		{name: "skipped on a terminal", terminal: true, stdin: "s\n", wantURL: "https://notes.example/mcp?debug=1"},
		{name: "overwritten on a terminal", terminal: true, stdin: "d\no\n", wantURL: "https://notes.example/mcp"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newHarness(t)
			path := filepath.Join(h.home, ".claude.json")
			h.copyFile("files/claude-edited.json", path)
			h.terminal = tc.terminal
			h.stdin = tc.stdin

			out, err := h.run("apply", "--yes")

			assert.Equal(t, tc.wantCode, exitCode(t, err))
			assert.Equal(t, tc.wantURL, gjson.Get(h.read(path), "mcpServers.notes.url").String())

			if tc.terminal {
				assert.Contains(t, out, "conflict ~/.claude.json mcpServers.notes (both changed)")
			}
		})
	}
}

func TestMergeLeavingTheManifestDeletesOwnedPaths(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	_, err := h.run("apply", "--yes")
	require.NoError(t, err)

	h.copyFile("manifests/empty.toml", filepath.Join(h.repo, "teire.toml"))
	require.NoError(t, os.RemoveAll(filepath.Join(h.repo, "claude")))

	out, err := h.run("apply", "--yes")
	require.NoError(t, err)
	assert.Contains(t, out, "    delete    mcpServers.notes\n")

	claudeJSON := h.read(filepath.Join(h.home, ".claude.json"))
	assert.False(t, gjson.Get(claudeJSON, "mcpServers.notes").Exists())
	assert.Equal(t, int64(3), gjson.Get(claudeJSON, "numStartups").Int())

	_, statErr := os.Lstat(filepath.Join(h.home, ".claude", "settings.json"))
	assert.ErrorIs(t, statErr, os.ErrNotExist)

	_, err = h.run("plan")
	assert.NoError(t, err)
}

func TestApplyRecordsTargetsAlreadyInPlace(t *testing.T) {
	t.Parallel()

	h := newHarness(t)
	_, err := h.run("apply", "--yes")
	require.NoError(t, err)
	require.NoError(t, os.Remove(h.env().StatePath()))

	out, err := h.run("apply", "--yes")
	require.NoError(t, err)
	assert.Equal(t, golden(t, "plan-in-sync.txt"), out)

	require.NoError(t, os.RemoveAll(filepath.Join(h.repo, "nvim")))

	out, err = h.run("plan")
	assert.Equal(t, 1, exitCode(t, err))
	assert.Contains(t, out, "(no longer managed)\n  remove    ~/.config/nvim\n")
}

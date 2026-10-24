package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"

	"github.com/fabioueno/teire/internal/strategy"
)

func TestLinkDiffShowsDifferentContent(t *testing.T) {
	t.Parallel()

	w := newWorld(t)
	w.copyFile("files/.zshrc", w.home(".zshrc"))
	c := strategy.LinkChange{Path: w.home(".zshrc"), Source: w.repo("zsh/.zshrc"), Reason: strategy.ReasonDifferentContent}

	diff, err := linkDiff(c)

	require.NoError(t, err)
	assert.Contains(t, diff, "@@ -1 +1 @@\n-export EDITOR=vim\n+export EDITOR=nvim\n")
}

func TestLinkDiffIsEmptyForOtherConflicts(t *testing.T) {
	t.Parallel()

	diff, err := linkDiff(strategy.LinkChange{Path: "/missing", Source: "/missing", Reason: "links to /elsewhere"})

	require.NoError(t, err)
	assert.Empty(t, diff)
}

func TestValueDiff(t *testing.T) {
	t.Parallel()

	w := newWorld(t)
	source := gjson.Get(w.read(w.repo("claude/claude.json")), "mcpServers.notes").Raw
	edited := gjson.Get(w.read("testdata/files/claude-edited.json"), "mcpServers.notes").Raw

	tests := []struct {
		name   string
		target string
		want   string
	}{
		{name: "both values indented", target: edited, want: "value.diff"},
		{name: "absent target shows only additions", want: "value-absent.diff"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			diff := valueDiff(strategy.Change{Path: "mcpServers.notes", Source: source, Target: tc.target})

			assert.Equal(t, golden(t, tc.want), diff)
		})
	}
}

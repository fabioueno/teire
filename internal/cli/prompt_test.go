package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/fabioueno/teire/internal/engine"
)

func TestConfirm(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  bool
	}{
		{input: "y\n", want: true},
		{input: "YES\n", want: true},
		{input: "n\n", want: false},
		{input: "\n", want: false},
		{input: "", want: false},
	}

	for _, tc := range tests {
		t.Run(strings.TrimSpace(tc.input), func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			got, err := newTerminal(strings.NewReader(tc.input), &out, "/Users/test").confirm("Apply these changes?")

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Equal(t, "Apply these changes? [y/N] ", out.String())
		})
	}
}

func TestResolve(t *testing.T) {
	t.Parallel()

	conflict := engine.Conflict{
		Item:   engine.Item{Path: "/Users/test/.claude.json"},
		Path:   "mcpServers.notes",
		Reason: "both changed",
		Diff:   "--- target\n+++ source\n",
	}

	tests := []struct {
		name     string
		input    string
		diff     string
		want     engine.Choice
		wantDiff string
	}{
		{name: "overwrite", input: "o\n", want: engine.Overwrite},
		{name: "skip", input: "s\n", want: engine.Skip},
		{name: "empty answer skips", input: "\n", want: engine.Skip},
		{name: "end of input skips", input: "", want: engine.Skip},
		{name: "diff, then overwrite", input: "d\no\n", want: engine.Overwrite, wantDiff: "--- target\n+++ source\n"},
		{name: "unknown answer asks again", input: "x\no\n", want: engine.Overwrite},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var out bytes.Buffer
			got, err := newTerminal(strings.NewReader(tc.input), &out, "/Users/test").resolve(conflict)

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
			assert.Contains(t, out.String(), "conflict ~/.claude.json mcpServers.notes (both changed)\n")
			assert.Contains(t, out.String(), tc.wantDiff)
		})
	}
}

func TestResolveWithoutADiff(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	conflict := engine.Conflict{Item: engine.Item{Path: "/Users/test/.zshrc"}, Reason: "links to /elsewhere"}

	_, err := newTerminal(strings.NewReader("d\ns\n"), &out, "/Users/test").resolve(conflict)

	require.NoError(t, err)
	assert.Contains(t, out.String(), "No content to compare.\n")
}

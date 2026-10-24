package engine

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"

	"github.com/pmezard/go-difflib/difflib"

	"github.com/fabioueno/teire/internal/strategy"
)

// linkDiff shows how a plain file differs from its source. Other link
// conflicts have no content to compare, and their reason says enough.
func linkDiff(c strategy.LinkChange) (string, error) {
	if c.Reason != strategy.ReasonDifferentContent {
		return "", nil
	}

	target, err := os.ReadFile(c.Path)
	if err != nil {
		return "", err
	}

	source, err := os.ReadFile(c.Source)
	if err != nil {
		return "", err
	}

	return unified(c.Path, c.Source, string(target), string(source))
}

func valueDiff(c strategy.Change) string {
	diff, err := unified("target "+c.Path, "source "+c.Path, indent(c.Target), indent(c.Source))
	if err != nil {
		return ""
	}

	return diff
}

func unified(from, to, a, b string) (string, error) {
	return difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
		A:        lines(a),
		B:        lines(b),
		FromFile: from,
		ToFile:   to,
		Context:  3,
	})
}

func indent(raw string) string {
	if raw == "" {
		return ""
	}

	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(raw), "", "  "); err != nil {
		return raw + "\n"
	}

	return buf.String() + "\n"
}

// lines splits text for the diff. An absent value has no lines at all, so the
// diff shows only additions instead of one empty line.
func lines(text string) []string {
	if text == "" {
		return nil
	}

	split := strings.SplitAfter(text, "\n")
	if split[len(split)-1] == "" {
		split = split[:len(split)-1]
	}

	return split
}

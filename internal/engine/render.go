package engine

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/fabioueno/teire/internal/manifest"
	"github.com/fabioueno/teire/internal/strategy"
)

const removedHeading = "(removed from the manifest)"

// Render writes the plan grouped by package, one line for each target that
// is not in sync, then a summary line.
func Render(w io.Writer, p Plan, home string) error {
	var b strings.Builder

	heading := ""
	unchanged := 0

	for _, item := range p.Items {
		if !item.Changed() {
			unchanged++
			continue
		}

		name := item.Package
		if item.Removed {
			name = removedHeading
		}

		if name != heading {
			heading = name
			fmt.Fprintln(&b, heading)
		}

		renderItem(&b, item, home)
	}

	fmt.Fprintln(&b, summary(p, unchanged))

	_, err := io.WriteString(w, b.String())

	return err
}

func renderItem(b *strings.Builder, item Item, home string) {
	path := ShortPath(item.Path, home)

	switch {
	case item.Missing && item.Removed:
		line(b, "release", path, "the target no longer exists")
		return
	case item.Missing:
		line(b, "missing", path, "the target does not exist; start its program once")
		return
	}

	switch item.Strategy {
	case manifest.Link, manifest.Folder:
		line(b, string(item.Link.Action), path, item.Link.Reason)
	case manifest.Merge:
		line(b, "merge", path, "")

		for _, c := range item.Merge {
			if c.Action != strategy.Unchanged {
				fmt.Fprintf(b, "    %-9s %s\n", c.Action, c.Path)
			}
		}
	}
}

func line(b *strings.Builder, action, path, reason string) {
	if reason == "" {
		fmt.Fprintf(b, "  %-9s %s\n", action, path)
		return
	}

	fmt.Fprintf(b, "  %-9s %s  (%s)\n", action, path, reason)
}

func summary(p Plan, unchanged int) string {
	changed := 0
	for _, item := range p.Items {
		if item.Changed() {
			changed++
		}
	}

	if changed == 0 {
		return fmt.Sprintf("Everything is in place (%d targets).", unchanged)
	}

	return fmt.Sprintf("%d to change, %d conflicts, %d unchanged.", changed, p.Conflicts(), unchanged)
}

// ShortPath writes a path below home with ~ in front.
func ShortPath(path, home string) string {
	rel, err := filepath.Rel(home, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return path
	}

	if rel == "." {
		return "~"
	}

	return "~/" + rel
}

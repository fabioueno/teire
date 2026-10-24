package strategy

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type LinkAction string

const ReasonDifferentContent = "has different content"

const (
	LinkUnchanged LinkAction = "unchanged"
	LinkCreate    LinkAction = "create"
	LinkReplace   LinkAction = "replace"
	LinkConflict  LinkAction = "conflict"
	LinkRemove    LinkAction = "remove"
	LinkRelease   LinkAction = "release"
)

// LinkChange is the plan for one link. Path is the target and Source is the
// absolute path the link points at. Reason says why a conflict or a replace
// was planned.
type LinkChange struct {
	Path   string
	Source string
	Action LinkAction
	Reason string
}

// PlanLink plans a link from path to source, a file or, with folder set, a
// whole folder in the repository at repo.
func PlanLink(path, source, repo string, folder bool) (LinkChange, error) {
	c := LinkChange{Path: path, Source: source}

	inside, err := ResolvesInto(filepath.Dir(path), repo)
	if err != nil {
		return LinkChange{}, err
	}

	if inside {
		c.Action, c.Reason = LinkConflict, "a parent folder links into the repository"
		return c, nil
	}

	info, err := os.Lstat(path)
	if errors.Is(err, fs.ErrNotExist) {
		c.Action = LinkCreate
		return c, nil
	}

	if err != nil {
		return LinkChange{}, fmt.Errorf("inspect %s: %w", path, err)
	}

	switch {
	case info.Mode().Type() == fs.ModeSymlink:
		return planOverLink(c)
	case info.IsDir() && !folder:
		c.Action, c.Reason = LinkConflict, "is a folder"
	case info.IsDir():
		c.Action, c.Reason = LinkConflict, "is a folder with its own files"
	case folder:
		c.Action, c.Reason = LinkConflict, "is a file"
	default:
		return planOverFile(c)
	}

	return c, nil
}

func planOverLink(c LinkChange) (LinkChange, error) {
	dest, err := os.Readlink(c.Path)
	if err != nil {
		return LinkChange{}, fmt.Errorf("read link %s: %w", c.Path, err)
	}

	if dest == c.Source {
		c.Action = LinkUnchanged
		return c, nil
	}

	if sameFile(c.Path, c.Source) {
		c.Action, c.Reason = LinkReplace, "links to the same file through "+dest
		return c, nil
	}

	// A dangling link holds no data, so replacing it loses nothing. Moving
	// files inside the repository leaves links like these behind.
	if _, err := os.Stat(c.Path); errors.Is(err, fs.ErrNotExist) {
		c.Action, c.Reason = LinkReplace, "links to a missing path "+dest
		return c, nil
	}

	c.Action, c.Reason = LinkConflict, "links to "+dest
	return c, nil
}

func planOverFile(c LinkChange) (LinkChange, error) {
	target, err := os.ReadFile(c.Path)
	if err != nil {
		return LinkChange{}, fmt.Errorf("read %s: %w", c.Path, err)
	}

	source, err := os.ReadFile(c.Source)
	if err != nil {
		return LinkChange{}, fmt.Errorf("read %s: %w", c.Source, err)
	}

	if bytes.Equal(target, source) {
		c.Action, c.Reason = LinkReplace, "has the same content"
		return c, nil
	}

	c.Action, c.Reason = LinkConflict, ReasonDifferentContent
	return c, nil
}

// PlanUnlink plans an entry that is no longer managed. teire removes the link
// only while it still points where teire made it point.
func PlanUnlink(path, recorded string) LinkChange {
	c := LinkChange{Path: path, Source: recorded, Action: LinkRelease}

	if dest, err := os.Readlink(path); err == nil && dest == recorded {
		c.Action = LinkRemove
	}

	return c
}

func sameFile(a, b string) bool {
	ia, err := os.Stat(a)
	if err != nil {
		return false
	}

	ib, err := os.Stat(b)
	if err != nil {
		return false
	}

	return os.SameFile(ia, ib)
}

// ResolvesInto reports whether dir, or the nearest folder above it that
// exists, resolves to a place inside repo. A link made there would land in
// the repository itself.
func ResolvesInto(dir, repo string) (bool, error) {
	root, err := filepath.EvalSymlinks(repo)
	if err != nil {
		return false, fmt.Errorf("resolve repository %s: %w", repo, err)
	}

	for {
		resolved, err := filepath.EvalSymlinks(dir)
		if err == nil {
			return resolved == root || strings.HasPrefix(resolved, root+string(filepath.Separator)), nil
		}

		if !errors.Is(err, fs.ErrNotExist) {
			return false, fmt.Errorf("resolve %s: %w", dir, err)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return false, nil
		}

		dir = parent
	}
}

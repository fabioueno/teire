package manifest

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"path"
	"path/filepath"
	"slices"
	"strings"
)

type Target struct {
	Package  string
	Source   string
	Path     string
	Strategy Strategy
	Own      []string
}

// Finder writes .DS_Store into any folder it shows, so a source folder can
// hold one that nobody meant to link.
const finderMetadata = ".DS_Store"

func Resolve(m Manifest, host string, repo fs.FS, home string) ([]Target, error) {
	packages := withHost(m, host)

	var targets []Target
	for _, name := range slices.Sorted(maps.Keys(packages)) {
		resolved, err := resolvePackage(name, packages[name], repo, home)
		if err != nil {
			return nil, fmt.Errorf("package %q: %w", name, err)
		}

		targets = append(targets, resolved...)
	}

	if err := checkUniquePaths(targets); err != nil {
		return nil, err
	}

	slices.SortFunc(targets, func(a, b Target) int {
		return strings.Compare(a.Path, b.Path)
	})

	return targets, nil
}

func resolvePackage(name string, p Package, repo fs.FS, home string) ([]Target, error) {
	source := p.Source
	if source == "" {
		source = name
	}

	source = path.Clean(source)

	if p.Target == "" {
		return nil, errors.New("target is required")
	}

	target, err := expandTarget(p.Target, home)
	if err != nil {
		return nil, err
	}

	info, err := fs.Stat(repo, source)
	if err != nil {
		return nil, fmt.Errorf("source folder %q: %w", source, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("source %q is not a folder", source)
	}

	strategy := p.Strategy
	if strategy == "" {
		strategy = Link
	}

	switch strategy {
	case Folder:
		if len(p.Files) > 0 {
			return nil, errors.New(`file entries are not valid with strategy "folder"`)
		}
		return []Target{{Package: name, Source: source, Path: target, Strategy: Folder}}, nil
	case Link:
		return resolveFiles(name, source, target, p.Files, repo, home)
	case Merge:
		return nil, errors.New(`strategy "merge" is only valid for a file entry`)
	default:
		return nil, fmt.Errorf("unknown strategy %q", strategy)
	}
}

func resolveFiles(name, source, target string, files map[string]File, repo fs.FS, home string) ([]Target, error) {
	pending := maps.Clone(files)

	var targets []Target
	err := fs.WalkDir(repo, source, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() == finderMetadata {
			return nil
		}

		rel := p[len(source)+1:]
		t := Target{
			Package:  name,
			Source:   p,
			Path:     filepath.Join(target, filepath.FromSlash(rel)),
			Strategy: Link,
		}

		if file, ok := pending[rel]; ok {
			delete(pending, rel)
			if t, err = applyFile(t, file, home); err != nil {
				return fmt.Errorf("file %q: %w", rel, err)
			}
		}

		targets = append(targets, t)
		return nil
	})

	if err != nil {
		return nil, err
	}

	if len(pending) > 0 {
		missing := slices.Sorted(maps.Keys(pending))
		return nil, fmt.Errorf("file %q: not found in %q", missing[0], source)
	}

	return targets, nil
}

func applyFile(t Target, file File, home string) (Target, error) {
	if file.Strategy != "" {
		t.Strategy = file.Strategy
	}

	if file.Target != "" {
		target, err := expandTarget(file.Target, home)
		if err != nil {
			return Target{}, err
		}

		t.Path = target
	}

	switch t.Strategy {
	case Merge:
		if len(file.Own) == 0 {
			return Target{}, errors.New(`strategy "merge" needs at least one path in own`)
		}

		t.Own = file.Own
	case Link:
		if len(file.Own) > 0 {
			return Target{}, errors.New(`own is only valid with strategy "merge"`)
		}
	case Folder:
		return Target{}, errors.New(`strategy "folder" is only valid for a package`)
	default:
		return Target{}, fmt.Errorf("unknown strategy %q", t.Strategy)
	}

	return t, nil
}

func checkUniquePaths(targets []Target) error {
	owners := make(map[string]Target, len(targets))
	for _, t := range targets {
		if other, ok := owners[t.Path]; ok {
			return fmt.Errorf("target %s comes from %s (package %q) and %s (package %q)",
				t.Path, other.Source, other.Package, t.Source, t.Package)
		}

		owners[t.Path] = t
	}

	return nil
}

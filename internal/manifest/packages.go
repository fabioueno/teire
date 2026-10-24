package manifest

import (
	"errors"
	"fmt"
	"maps"
	"path"
	"slices"
)

// ResolvedPackage is a package with its defaults filled in and its target
// expanded, before its source folder is read.
type ResolvedPackage struct {
	Name     string
	Source   string
	Target   string
	Strategy Strategy
}

func Packages(m Manifest, host, home string) ([]ResolvedPackage, error) {
	packages := withHost(m, host)

	resolved := make([]ResolvedPackage, 0, len(packages))
	for _, name := range slices.Sorted(maps.Keys(packages)) {
		pkg, err := resolveHead(name, packages[name], home)
		if err != nil {
			return nil, fmt.Errorf("package %q: %w", name, err)
		}

		resolved = append(resolved, pkg)
	}

	return resolved, nil
}

func resolveHead(name string, p Package, home string) (ResolvedPackage, error) {
	source := p.Source
	if source == "" {
		source = name
	}

	if p.Target == "" {
		return ResolvedPackage{}, errors.New("target is required")
	}

	target, err := expandTarget(p.Target, home)
	if err != nil {
		return ResolvedPackage{}, err
	}

	strategy := p.Strategy
	if strategy == "" {
		strategy = Link
	}

	return ResolvedPackage{Name: name, Source: path.Clean(source), Target: target, Strategy: strategy}, nil
}

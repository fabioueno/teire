package manifest

import (
	"fmt"
	"io/fs"
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

func Packages(m Manifest, host string, repo fs.FS, home string) ([]ResolvedPackage, error) {
	packages, err := withHost(m, host, repo)
	if err != nil {
		return nil, err
	}

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

	defaults := defaultsFor(name)

	target := p.Target
	if target == "" {
		target = defaults.Target
	}

	expanded, err := expandTarget(target, home)
	if err != nil {
		return ResolvedPackage{}, err
	}

	strategy := p.Strategy
	if strategy == "" {
		strategy = defaults.Strategy
	}

	return ResolvedPackage{Name: name, Source: path.Clean(source), Target: expanded, Strategy: strategy}, nil
}

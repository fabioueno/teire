package manifest

import (
	"fmt"
	"io/fs"
	"maps"
	"slices"
)

// withHost gives the packages for one host: the declared ones, the known
// folders found in the repository, then the host section over both. Ignoring a
// name stops discovery only, so a declared package cannot also be ignored.
func withHost(m Manifest, host string, repo fs.FS) (map[string]Package, error) {
	packages := maps.Clone(m.Packages)
	if packages == nil {
		packages = map[string]Package{}
	}

	ignored := ignoredNames(m, host)
	if err := checkNotIgnored(packages, m.Hosts[host].Packages, ignored); err != nil {
		return nil, err
	}

	discover(packages, repo, ignored)

	for name, over := range m.Hosts[host].Packages {
		base, ok := packages[name]
		if !ok {
			packages[name] = over
			continue
		}

		packages[name] = overlayPackage(base, over)
	}

	return packages, nil
}

func ignoredNames(m Manifest, host string) map[string]bool {
	names := slices.Concat(m.Ignore, m.Hosts[host].Ignore)
	ignored := make(map[string]bool, len(names))
	for _, name := range names {
		ignored[name] = true
	}

	return ignored
}

func checkNotIgnored(base, host map[string]Package, ignored map[string]bool) error {
	for _, name := range slices.Sorted(maps.Keys(ignored)) {
		_, inBase := base[name]
		_, inHost := host[name]
		if inBase || inHost {
			return fmt.Errorf("package %q is both declared and ignored", name)
		}
	}

	return nil
}

func overlayPackage(base, over Package) Package {
	if over.Source != "" {
		base.Source = over.Source
	}
	if over.Target != "" {
		base.Target = over.Target
	}
	if over.Strategy != "" {
		base.Strategy = over.Strategy
	}

	files := maps.Clone(base.Files)
	for rel, overFile := range over.Files {
		if files == nil {
			files = map[string]File{}
		}

		files[rel] = overlayFile(files[rel], overFile)
	}

	base.Files = files

	return base
}

func overlayFile(base, over File) File {
	if over.Strategy != "" {
		base.Strategy = over.Strategy
	}
	if over.Target != "" {
		base.Target = over.Target
	}
	if over.Own != nil {
		base.Own = over.Own
	}

	return base
}

package manifest

import "maps"

func withHost(m Manifest, host string) map[string]Package {
	packages := maps.Clone(m.Packages)
	if packages == nil {
		packages = map[string]Package{}
	}

	for name, over := range m.Hosts[host].Packages {
		base, ok := packages[name]
		if !ok {
			packages[name] = over
			continue
		}

		packages[name] = overlayPackage(base, over)
	}

	return packages
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

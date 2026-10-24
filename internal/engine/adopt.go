package engine

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/fabioueno/teire/internal/env"
	"github.com/fabioueno/teire/internal/fsx"
	"github.com/fabioueno/teire/internal/manifest"
	"github.com/fabioueno/teire/internal/state"
	"github.com/fabioueno/teire/internal/strategy"
)

// Adopt moves the file at path into the source folder of the link package
// whose target covers it, links it back, and returns where the file went.
func Adopt(e env.Env, packages []manifest.ResolvedPackage, st state.State, path string) (string, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return "", fmt.Errorf("adopt %s: %w", path, err)
	}

	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("adopt %s: only a plain file can be adopted", path)
	}

	pkg, rel, ok := coveringPackage(packages, path)
	if !ok {
		return "", fmt.Errorf("adopt %s: no link package has a target that covers it", path)
	}

	inside, err := strategy.ResolvesInto(filepath.Dir(path), e.Repo)
	if err != nil {
		return "", err
	}

	if inside {
		return "", fmt.Errorf("adopt %s: a parent folder links into the repository", path)
	}

	dest := filepath.Join(e.Repo, filepath.FromSlash(pkg.Source), rel)
	if _, err := os.Lstat(dest); !errors.Is(err, fs.ErrNotExist) {
		return "", fmt.Errorf("adopt %s: %s already exists", path, dest)
	}

	if err := fsx.Move(path, dest); err != nil {
		return "", err
	}

	if err := fsx.Symlink(dest, path); err != nil {
		return "", err
	}

	st.Targets[path] = state.Entry{Strategy: manifest.Link, Source: dest}

	return dest, nil
}

// coveringPackage picks the link package with the deepest target that holds
// path, so ~/.config/zed wins over ~ for ~/.config/zed/settings.json.
func coveringPackage(packages []manifest.ResolvedPackage, path string) (manifest.ResolvedPackage, string, bool) {
	var best manifest.ResolvedPackage
	bestRel, found := "", false

	for _, pkg := range packages {
		if pkg.Strategy != manifest.Link {
			continue
		}

		rel, err := filepath.Rel(pkg.Target, path)
		if err != nil || rel == "." || strings.HasPrefix(rel, "..") {
			continue
		}

		if !found || len(pkg.Target) > len(best.Target) {
			best, bestRel, found = pkg, rel, true
		}
	}

	return best, bestRel, found
}

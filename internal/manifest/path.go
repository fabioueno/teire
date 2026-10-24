package manifest

import (
	"fmt"
	"path/filepath"
	"strings"
)

func expandTarget(target, home string) (string, error) {
	switch {
	case target == "~":
		target = home
	case strings.HasPrefix(target, "~/"):
		target = filepath.Join(home, target[2:])
	}

	if !filepath.IsAbs(target) {
		return "", fmt.Errorf("target %q is not absolute and does not start with ~/", target)
	}

	return filepath.Clean(target), nil
}

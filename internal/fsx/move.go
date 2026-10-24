package fsx

import (
	"fmt"
	"os"
	"path/filepath"
)

// Move renames path to dest, creating the folders above dest.
func Move(path, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("create folder for %s: %w", dest, err)
	}

	if err := os.Rename(path, dest); err != nil {
		return fmt.Errorf("move %s: %w", path, err)
	}

	return nil
}

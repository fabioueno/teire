package fsx

import (
	"fmt"
	"os"
	"path/filepath"
)

// Backup moves path into root, under its own absolute path, and returns where
// it went.
func Backup(path, root string) (string, error) {
	dest := filepath.Join(root, path)
	if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
		return "", fmt.Errorf("create backup folder for %s: %w", path, err)
	}

	if err := os.Rename(path, dest); err != nil {
		return "", fmt.Errorf("back up %s: %w", path, err)
	}

	return dest, nil
}

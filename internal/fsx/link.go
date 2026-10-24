package fsx

import (
	"fmt"
	"os"
	"path/filepath"
)

func Symlink(source, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create folder for %s: %w", path, err)
	}

	if err := os.Symlink(source, path); err != nil {
		return fmt.Errorf("link %s: %w", path, err)
	}

	return nil
}

func RemoveLink(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("remove %s: %w", path, err)
	}

	if info.Mode().Type() != os.ModeSymlink {
		return fmt.Errorf("remove %s: not a symlink", path)
	}

	return os.Remove(path)
}

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

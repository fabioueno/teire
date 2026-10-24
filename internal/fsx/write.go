package fsx

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func MkdirAll(path string, perm fs.FileMode) error {
	return os.MkdirAll(path, perm)
}

func WriteFileAtomic(path string, data []byte, perm fs.FileMode) (err error) {
	// A rename over a symlink replaces the link with a plain file, so write
	// to the file the link points at.
	real, err := filepath.EvalSymlinks(path)

	switch {
	case err == nil:
		path = real
	case !errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("resolve %s: %w", path, err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".teire-*")
	if err != nil {
		return fmt.Errorf("create temporary file for %s: %w", path, err)
	}

	defer func() {
		if err != nil {
			_ = os.Remove(tmp.Name())
		}
	}()

	if err = writeAndClose(tmp, data, perm); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}

	if err = os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}

	return nil
}

func writeAndClose(f *os.File, data []byte, perm fs.FileMode) error {
	if _, err := f.Write(data); err != nil {
		_ = f.Close()
		return err
	}

	if err := f.Chmod(perm); err != nil {
		_ = f.Close()
		return err
	}

	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}

	return f.Close()
}

package fsx

import (
	"fmt"
	"os"
)

// RemoveAll deletes path and everything below it. A path that does not exist
// is not an error.
func RemoveAll(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("remove %s: %w", path, err)
	}

	return nil
}

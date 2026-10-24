package strategy

import (
	"bytes"
	"fmt"
	"os"

	"github.com/tidwall/sjson"

	"github.com/fabioueno/teire/internal/fsx"
)

// PullMerge writes the target value of each drifted change into the source
// file, and keeps every other byte of the file.
func PullMerge(source string, changes []Change) error {
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("stat %s: %w", source, err)
	}

	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("read %s: %w", source, err)
	}

	out := data
	for _, c := range changes {
		out, err = sjson.SetRawBytes(out, c.Path, compact(c.Target))
		if err != nil {
			return fmt.Errorf("pull into %s: path %q: %w", source, c.Path, err)
		}
	}

	if bytes.Equal(out, data) {
		return nil
	}

	return fsx.WriteFileAtomic(source, out, info.Mode().Perm())
}

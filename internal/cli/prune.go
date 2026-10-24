package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/spf13/cobra"

	"github.com/fabioueno/teire/internal/engine"
	"github.com/fabioueno/teire/internal/env"
	"github.com/fabioueno/teire/internal/fsx"
)

const defaultKeep = 10

// newPrune does not look for a repository: it needs only the state folder.
func newPrune(d deps) *cobra.Command {
	var keep int

	cmd := &cobra.Command{
		Use:   "prune [--keep N]",
		Short: "Remove all but the newest backups",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if keep < 0 {
				return errors.New("--keep must not be negative")
			}

			e, err := d.env()
			if err != nil {
				return err
			}

			return runPrune(cmd, e, keep)
		},
	}

	cmd.Flags().IntVar(&keep, "keep", defaultKeep, "number of newest backups to keep")

	return cmd
}

func runPrune(cmd *cobra.Command, e env.Env, keep int) error {
	names, err := backupNames(e.BackupsDir())
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()

	if len(names) <= keep {
		_, err := fmt.Fprintf(out, "Nothing to prune (%s, keeping %d).\n", count(len(names)), keep)

		return err
	}

	stale := names[:len(names)-keep]

	for _, name := range stale {
		path := filepath.Join(e.BackupsDir(), name)
		if err := fsx.RemoveAll(path); err != nil {
			return err
		}

		if _, err := fmt.Fprintf(out, "  removed   %s\n", engine.ShortPath(path, e.Home)); err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(out, "Removed %s, kept %d.\n", count(len(stale)), keep)

	return err
}

// backupNames lists the backup folders oldest first. The timestamp format
// sorts by time, and any other entry is not a backup, so it stays.
func backupNames(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("read backups: %w", err)
	}

	var names []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if _, err := time.Parse(env.BackupLayout, entry.Name()); err == nil {
			names = append(names, entry.Name())
		}
	}

	slices.Sort(names)

	return names, nil
}

func count(n int) string {
	if n == 1 {
		return "1 backup"
	}

	return fmt.Sprintf("%d backups", n)
}

package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/fabioueno/teire/internal/engine"
	"github.com/fabioueno/teire/internal/manifest"
)

func newAdopt(d deps, opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "adopt <path>",
		Short: "Move a file into its package folder and link it back",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			w, err := load(d, opts)
			if err != nil {
				return err
			}

			packages, err := manifest.Packages(w.manifest, w.host, os.DirFS(w.env.Repo), w.env.Home)
			if err != nil {
				return err
			}

			path, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}

			dest, err := engine.Adopt(w.env, packages, w.state, path)
			if err != nil {
				return err
			}

			if err := w.state.Save(w.env.StatePath()); err != nil {
				return err
			}

			rel, err := filepath.Rel(w.env.Repo, dest)
			if err != nil {
				rel = dest
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "Adopted %s into %s.\n", engine.ShortPath(path, w.env.Home), rel)

			return err
		},
	}
}

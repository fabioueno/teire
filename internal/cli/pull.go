package cli

import (
	"errors"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/fabioueno/teire/internal/engine"
)

func newPull(d deps, opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "pull [path...]",
		Short: "Bring owned keys changed in merge targets back into the repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPull(cmd, d, opts, args)
		},
	}
}

func runPull(cmd *cobra.Command, d deps, opts *options, args []string) error {
	w, err := load(d, opts)
	if err != nil {
		return err
	}

	p, err := w.plan()
	if err != nil {
		return err
	}

	paths := make([]string, 0, len(args))

	for _, arg := range args {
		path, err := filepath.Abs(arg)
		if err != nil {
			return err
		}

		paths = append(paths, path)
	}

	next, result, pullErr := engine.Pull(w.env, p, w.state, paths, d.now())
	if err := next.Save(w.env.StatePath()); err != nil {
		return errors.Join(pullErr, err)
	}

	if pullErr != nil {
		return pullErr
	}

	return engine.RenderPull(cmd.OutOrStdout(), result, w.env.Home)
}

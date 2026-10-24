// Package cli holds the teire commands. It is the only package that imports
// cobra, so the other packages stay free of command-line types.
package cli

import "github.com/spf13/cobra"

func NewRoot() *cobra.Command {
	return newRoot(osDeps())
}

func newRoot(d deps) *cobra.Command {
	opts := &options{}

	root := &cobra.Command{
		Use:           "teire",
		Short:         "Keep a machine's dotfiles in the shape a repository declares",
		Args:          cobra.NoArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runPlan(cmd, d, opts)
		},
	}

	flags := root.PersistentFlags()
	flags.StringVar(&opts.repo, "repo", "", "dotfiles repository (default: the folder above with teire.toml)")
	flags.StringVar(&opts.host, "host", "", "host name for the manifest's host sections")

	root.AddCommand(
		newVersion(),
		newInit(d, opts),
		newPlan(d, opts),
		newApply(d, opts),
		newAdopt(d, opts),
		newPrune(d),
		newPull(d, opts),
	)

	return root
}

// Package cli holds the teire commands. It is the only package that imports
// cobra, so the other packages stay free of command-line types.
package cli

import "github.com/spf13/cobra"

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "teire",
		Short:         "Keep a machine's dotfiles in the shape a repository declares",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	root.AddCommand(newVersion())
	return root
}

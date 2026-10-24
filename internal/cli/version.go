package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Set at build time by the Makefile through -ldflags -X.
var version = "dev"

func newVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the teire version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), version)
			return err
		},
	}
}

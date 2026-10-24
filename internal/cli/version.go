package cli

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
)

// Set at build time through -ldflags -X, by the Makefile and the Homebrew
// formula.
var version = "dev"

func newVersion() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the teire version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			info, ok := debug.ReadBuildInfo()
			_, err := fmt.Fprintln(cmd.OutOrStdout(), resolveVersion(version, info, ok))

			return err
		},
	}
}

// resolveVersion prefers the stamped version. go install stamps nothing, but
// Go records the module version in the binary, so that version comes next.
func resolveVersion(stamped string, info *debug.BuildInfo, ok bool) string {
	if stamped != "dev" {
		return stamped
	}

	// A build from a checkout records "(devel)", which says nothing more.
	if !ok || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return stamped
	}

	return strings.TrimPrefix(info.Main.Version, "v")
}

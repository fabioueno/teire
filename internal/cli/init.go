package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/fabioueno/teire/internal/engine"
	"github.com/fabioueno/teire/internal/fsx"
)

func newInit(d deps, opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "init --host <name>",
		Short: "Name this machine for the manifest's host sections",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.host == "" {
				return errors.New("init needs --host <name>")
			}

			e, err := d.env()
			if err != nil {
				return err
			}

			if err := fsx.MkdirAll(e.ConfigDir, 0o755); err != nil {
				return fmt.Errorf("create config folder: %w", err)
			}

			if err := fsx.WriteFileAtomic(e.HostFile(), []byte(opts.host+"\n"), 0o644); err != nil {
				return err
			}

			_, err = fmt.Fprintf(cmd.OutOrStdout(), "This machine is %q (saved in %s).\n", opts.host, engine.ShortPath(e.HostFile(), e.Home))

			return err
		},
	}
}

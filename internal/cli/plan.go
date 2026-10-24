package cli

import (
	"github.com/spf13/cobra"

	"github.com/fabioueno/teire/internal/engine"
)

func newPlan(d deps, opts *options) *cobra.Command {
	return &cobra.Command{
		Use:   "plan",
		Short: "Show what apply would change, without writing (exit 1: changes, 2: conflicts)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runPlan(cmd, d, opts)
		},
	}
}

func runPlan(cmd *cobra.Command, d deps, opts *options) error {
	w, err := load(d, opts)
	if err != nil {
		return err
	}

	p, err := w.plan()
	if err != nil {
		return err
	}

	if err := engine.Render(cmd.OutOrStdout(), p, w.env.Home); err != nil {
		return err
	}

	switch {
	case p.Conflicts() > 0:
		return &exitError{code: 2}
	case p.Changed():
		return &exitError{code: 1}
	default:
		return nil
	}
}

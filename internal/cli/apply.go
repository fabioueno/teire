package cli

import (
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"github.com/fabioueno/teire/internal/engine"
)

func newApply(d deps, opts *options) *cobra.Command {
	var yes bool

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Write the changes that plan shows",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runApply(cmd, d, opts, yes)
		},
	}

	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "apply without asking first")

	return cmd
}

func runApply(cmd *cobra.Command, d deps, opts *options, yes bool) error {
	w, err := load(d, opts)
	if err != nil {
		return err
	}

	p, err := w.plan()
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	if err := engine.Render(out, p, w.env.Home); err != nil {
		return err
	}

	if !p.Changed() {
		return nil
	}

	interactive := d.isTerminal()
	t := newTerminal(d.stdin, out, w.env.Home)

	if !yes {
		if !interactive {
			return errors.New("stdin is not a terminal; pass --yes to apply without a question")
		}

		ok, err := t.confirm("Apply these changes?")
		if err != nil {
			return err
		}

		if !ok {
			_, err := fmt.Fprintln(out, "Nothing applied.")
			return err
		}
	}

	resolve := skipConflict
	if interactive {
		resolve = t.resolve
	}

	next, result, applyErr := engine.Apply(w.env, p, w.state, resolve, d.now())
	if err := next.Save(w.env.StatePath()); err != nil {
		return errors.Join(applyErr, err)
	}

	if applyErr != nil {
		return applyErr
	}

	if err := report(out, result, w.env.Home); err != nil {
		return err
	}

	if result.Skipped > 0 && !interactive {
		return &exitError{code: 2, err: fmt.Errorf("%d conflicts left; run teire apply on a terminal to resolve them", result.Skipped)}
	}

	return nil
}

func skipConflict(engine.Conflict) (engine.Choice, error) {
	return engine.Skip, nil
}

func report(w io.Writer, r engine.Result, home string) error {
	var b strings.Builder

	for _, path := range r.Missing {
		fmt.Fprintf(&b, "%s does not exist yet; start its program once, then run apply again.\n", engine.ShortPath(path, home))
	}

	for _, path := range slices.Sorted(maps.Keys(r.NotKept)) {
		fmt.Fprintf(&b, "%s: %s did not stay after the write; another program may have written the file.\n",
			engine.ShortPath(path, home), strings.Join(r.NotKept[path], ", "))
	}

	if r.Skipped > 0 {
		fmt.Fprintf(&b, "%d conflicts skipped.\n", r.Skipped)
	}

	b.WriteString("Applied.\n")

	_, err := io.WriteString(w, b.String())

	return err
}

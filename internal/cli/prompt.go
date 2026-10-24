package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/fabioueno/teire/internal/engine"
)

type terminal struct {
	in   *bufio.Reader
	out  io.Writer
	home string
}

func newTerminal(in io.Reader, out io.Writer, home string) terminal {
	return terminal{in: bufio.NewReader(in), out: out, home: home}
}

func (t terminal) confirm(question string) (bool, error) {
	if _, err := fmt.Fprintf(t.out, "%s [y/N] ", question); err != nil {
		return false, err
	}

	answer, err := t.readLine()
	if err != nil {
		return false, err
	}

	return answer == "y" || answer == "yes", nil
}

func (t terminal) resolve(c engine.Conflict) (engine.Choice, error) {
	target := engine.ShortPath(c.Item.Path, t.home)
	if c.Path != "" {
		target += " " + c.Path
	}

	if _, err := fmt.Fprintf(t.out, "\nconflict %s (%s)\n", target, c.Reason); err != nil {
		return engine.Skip, err
	}

	for {
		if _, err := fmt.Fprint(t.out, "[d]iff, [o]verwrite, [s]kip? "); err != nil {
			return engine.Skip, err
		}

		answer, err := t.readLine()
		if err != nil {
			return engine.Skip, err
		}

		switch answer {
		case "d":
			if err := t.showDiff(c.Diff); err != nil {
				return engine.Skip, err
			}
		case "o":
			return engine.Overwrite, nil
		case "s", "":
			return engine.Skip, nil
		}
	}
}

func (t terminal) showDiff(diff string) error {
	if diff == "" {
		diff = "No content to compare.\n"
	}

	_, err := io.WriteString(t.out, diff)

	return err
}

// readLine gives the answer in lower case. The end of the input counts as an
// empty answer, so a closed stdin takes the safe default.
func (t terminal) readLine() (string, error) {
	line, err := t.in.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	return strings.ToLower(strings.TrimSpace(line)), nil
}

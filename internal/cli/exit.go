package cli

import (
	"errors"
	"fmt"
)

// exitError ends teire with a code. Without a message it is silent, because
// plan uses the code itself to say whether the machine is in sync.
type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string {
	if e.err == nil {
		return fmt.Sprintf("exit status %d", e.code)
	}

	return e.err.Error()
}

func (e *exitError) Unwrap() error {
	return e.err
}

// ExitCode gives the process exit code for err, and whether main should
// print the error.
func ExitCode(err error) (int, bool) {
	if exit, ok := errors.AsType[*exitError](err); ok {
		return exit.code, exit.err != nil
	}

	return 1, true
}

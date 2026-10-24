// Command teire keeps a machine's dotfiles in the shape that a repository
// declares.
package main

import (
	"fmt"
	"os"

	"github.com/fabioueno/teire/internal/cli"
)

func main() {
	if err := cli.NewRoot().Execute(); err != nil {
		code, show := cli.ExitCode(err)
		if show {
			fmt.Fprintln(os.Stderr, "teire:", err)
		}

		os.Exit(code)
	}
}

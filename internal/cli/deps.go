package cli

import (
	"io"
	"os"
	"time"

	"golang.org/x/term"

	"github.com/fabioueno/teire/internal/env"
	"github.com/fabioueno/teire/internal/host"
)

// deps holds everything the commands take from the machine, so tests can run
// them against temporary folders and scripted answers.
type deps struct {
	env           func() (env.Env, error)
	getwd         func() (string, error)
	localHostName func() (string, error)
	stdin         io.Reader
	isTerminal    func() bool
	now           func() time.Time
}

type options struct {
	repo string
	host string
}

func osDeps() deps {
	return deps{
		env:           env.FromOS,
		getwd:         os.Getwd,
		localHostName: host.LocalHostName,
		stdin:         os.Stdin,
		isTerminal:    func() bool { return term.IsTerminal(int(os.Stdin.Fd())) },
		now:           time.Now,
	}
}

// Package host finds the name of the machine that teire runs on.
package host

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"
)

func Name(flag, file string, localHostName func() (string, error)) (string, error) {
	if flag != "" {
		return flag, nil
	}

	data, err := os.ReadFile(file)

	switch {
	case err == nil:
		name := strings.TrimSpace(string(data))
		if name == "" {
			return "", fmt.Errorf("host file %s is empty", file)
		}

		return name, nil
	case !errors.Is(err, fs.ErrNotExist):
		return "", fmt.Errorf("read host file: %w", err)
	}

	name, err := localHostName()
	if err != nil {
		return "", fmt.Errorf("get local host name: %w", err)
	}

	return strings.TrimSpace(name), nil
}

func LocalHostName() (string, error) {
	out, err := exec.Command("scutil", "--get", "LocalHostName").Output()
	return string(out), err
}

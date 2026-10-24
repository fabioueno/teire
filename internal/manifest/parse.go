// Package manifest reads teire.toml and resolves it into the targets for one
// host.
package manifest

import (
	"fmt"
	"io"
	"strings"

	"github.com/BurntSushi/toml"
)

type Strategy string

const (
	Link   Strategy = "link"
	Folder Strategy = "folder"
	Merge  Strategy = "merge"
)

type Manifest struct {
	Packages map[string]Package `toml:"package"`
	Hosts    map[string]Host    `toml:"host"`
}

type Package struct {
	Source   string          `toml:"source"`
	Target   string          `toml:"target"`
	Strategy Strategy        `toml:"strategy"`
	Files    map[string]File `toml:"file"`
}

type File struct {
	Strategy Strategy `toml:"strategy"`
	Target   string   `toml:"target"`
	Own      []string `toml:"own"`
}

type Host struct {
	Packages map[string]Package `toml:"package"`
}

func Parse(r io.Reader) (Manifest, error) {
	var m Manifest
	md, err := toml.NewDecoder(r).Decode(&m)
	if err != nil {
		return Manifest{}, fmt.Errorf("parse manifest: %w", err)
	}

	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		keys := make([]string, len(undecoded))
		for i, key := range undecoded {
			keys[i] = key.String()
		}

		return Manifest{}, fmt.Errorf("parse manifest: unknown keys: %s", strings.Join(keys, ", "))
	}

	return m, nil
}

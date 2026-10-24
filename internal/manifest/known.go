package manifest

import (
	"io/fs"
	"path"
)

// known holds the defaults for common packages. A known folder in the
// repository needs no entry in teire.toml. Any other package must be
// declared, and defaults to ~/.config/<name> with the files strategy.
var known = map[string]Package{
	"alacritty": {Target: "~/.config/alacritty", Strategy: Files},
	"bash":      {Target: "~", Strategy: Files},
	"claude":    {Target: "~/.claude", Strategy: Files},
	"fish":      {Target: "~/.config/fish", Strategy: Files},
	"gh":        {Target: "~/.config/gh", Strategy: Files},
	"ghostty":   {Target: "~/.config/ghostty", Strategy: Files},
	"git":       {Target: "~", Strategy: Files},
	"helix":     {Target: "~/.config/helix", Strategy: Files},
	"kitty":     {Target: "~/.config/kitty", Strategy: Files},
	"mise":      {Target: "~/.config/mise", Strategy: Files},
	"nvim":      {Target: "~/.config/nvim", Strategy: Folder},
	"starship":  {Target: "~/.config", Strategy: Files},
	"tmux":      {Target: "~/.config/tmux", Strategy: Files},
	"vim":       {Target: "~", Strategy: Files},
	"wezterm":   {Target: "~/.config/wezterm", Strategy: Files},
	"zed":       {Target: "~/.config/zed", Strategy: Files},
	"zsh":       {Target: "~", Strategy: Files},
}

func defaultsFor(name string) Package {
	if p, ok := known[name]; ok {
		return p
	}

	return Package{Target: "~/.config/" + name, Strategy: Files}
}

// discover adds a package for each known folder in the repository that no
// declared package uses as its source.
func discover(packages map[string]Package, repo fs.FS) {
	used := make(map[string]bool, len(packages))
	for name, p := range packages {
		source := p.Source
		if source == "" {
			source = name
		}

		used[path.Clean(source)] = true
	}

	for name := range known {
		if _, declared := packages[name]; declared || used[name] {
			continue
		}

		if info, err := fs.Stat(repo, name); err == nil && info.IsDir() {
			packages[name] = Package{}
		}
	}
}

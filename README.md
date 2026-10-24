<div align="center">
  <h1>teire (手入れ)</h1>

  <p>Keeps a machine's dotfiles in the shape that a repository declares.</p>
</div>

---

Teire links the files in a dotfiles repository into your home folder, merges
the keys you own into config files that other programs also write, and records
each apply, so it can tell your changes from a program's changes. It is one
static binary with a plan and apply workflow: `teire plan` shows what would
change and writes nothing, and `teire apply` makes it so.

## Why another dotfiles tool

Good tools already exist, and each of them makes a tradeoff that Teire makes
differently.

| Tool               | What it does well                                 | Where Teire differs                                                                  |
| ------------------ | ------------------------------------------------- | ------------------------------------------------------------------------------------ |
| [GNU Stow][stow]   | Links a package folder into a target, no setup.   | Stow keeps no record, so it cannot tell who changed a file.                          |
| [chezmoi][chezmoi] | Templates, secrets, scripts, a state of applies.  | File names carry the behavior, and edits to part of a file do not track their keys.  |
| [yadm][yadm]       | Your home folder is the git work tree.            | Files live in place; there is no link, merge or plan step.                           |
| [dotbot][dotbot]   | A YAML list of links, run from the repository.    | It keeps no state, and it cannot own part of a file.                                 |
| [home-manager][hm] | The whole environment, declared in Nix.           | Files link into the read-only Nix store, so programs cannot save their own settings. |

Teire exists for three cases that these tools handle poorly together:

- **Files that a program also writes.** Claude Code rewrites `~/.claude.json`
  every few minutes with account data and caches, and keeps your MCP servers
  in the same file. Teire owns only the keys you list, writes only those, and
  removes a key you stop owning, because its state records what it set.
- **Folders that a program fills.** Linking `~/.claude` as one folder would
  send transcripts and history into your repository. Teire's default links
  each file into a real folder, and links a whole folder only where only you
  write to it, such as `~/.config/nvim`.
- **Conflicts with a reason.** With a record of the last apply, Teire knows
  whether the repository changed, the target changed, or both. Only the last
  case asks you.

## Install

With Homebrew:

```sh
brew install fabioueno/tap/teire
```

Or with Go 1.27 or later:

```sh
go install github.com/fabioueno/teire/cmd/teire@latest
```

## Quick start

A package is a folder in your repository whose layout mirrors its target.
Teire knows the target of common packages, so a repository like this needs no
configuration at all:

```text
dotfiles/
├── git/.gitconfig          → ~/.gitconfig
├── nvim/                   → ~/.config/nvim, linked as one folder
├── starship/starship.toml  → ~/.config/starship.toml
└── zsh/.zshrc              → ~/.zshrc
```

```sh
cd dotfiles
teire plan     # shows what would change; exit 1 for changes, 2 for conflicts
teire apply    # asks once, then links everything
```

Add a `teire.toml` at the root only for what differs from the defaults: a
package Teire does not know, a known folder to leave unlinked, a file that needs
`merge`, or a machine with its own targets.

```toml
# Keep the zed folder in the repository, but do not link it.
ignore = ["zed"]

# Own one MCP server in ~/.claude.json, and leave every other key alone.
[package.claude.file."claude.json"]
strategy = "merge"
target   = "~/.claude.json"
own      = ["mcpServers.notes"]

# On the machine named "work", Claude Code runs from another config folder.
[host.work.package.claude]
target = "~/.claude-work"

[host.work.package.claude.file."claude.json"]
target = "~/.claude-work/.claude.json"
```

## Commands

| Command              | What it does                                                        |
| -------------------- | ------------------------------------------------------------------- |
| `teire plan`         | Shows each change and conflict, and writes nothing.                 |
| `teire apply`        | Shows the plan, asks once, then writes. `--yes` skips the question. |
| `teire pull [path]`  | Brings owned keys changed in a `merge` target back into the repo.   |
| `teire adopt <path>` | Moves an existing file into its package and links it back.          |
| `teire init --host`  | Names this machine, for the host sections in `teire.toml`.          |
| `teire prune`        | Removes all but the 10 newest backups. `--keep N` changes the 10.   |

Every overwrite first saves the old target under
`~/.local/state/teire/backups/`, and `teire prune` removes the old ones.

## Status

Early development. The `files`, `folder` and `merge` strategies work on
macOS. macOS settings and Homebrew packages are planned.

## Documentation

| Document                       | Description                                          |
| ------------------------------ | ---------------------------------------------------- |
| [Architecture][architecture]   | The manifest, the strategies, state, and the code.   |
| [Tradeoffs][tradeoffs]         | Alternatives turned down, and why.                   |
| [Deferred decisions][deferred] | Questions still open.                                |
| [Contributing][contributing]   | Development setup and conventions.                   |

## License

[MIT][license]

[architecture]: docs/architecture/overview.md
[chezmoi]: https://www.chezmoi.io
[contributing]: CONTRIBUTING.md
[deferred]: docs/design/deferred-decisions.md
[dotbot]: https://github.com/anishathalye/dotbot
[hm]: https://github.com/nix-community/home-manager
[license]: LICENSE
[stow]: https://www.gnu.org/software/stow/
[tradeoffs]: docs/design/tradeoffs.md
[yadm]: https://yadm.io

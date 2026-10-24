# Tradeoffs

The alternatives turned down for this project, and what would justify revisiting
each.

## A tool of its own, not chezmoi

chezmoi links files, edits parts of files with modify templates, records the
last apply, and runs scripts for macOS settings. It was turned down for two
reasons. File names carry its behavior. And a modify template does not know
which keys it set last time, so a key removed from the source stays in the
target.

Revisit when the upkeep of Teire costs more than these two gaps.

## Go, not TypeScript or Rust

The first run on a new machine happens before Node is installed, so the tool
must be one static binary. TypeScript with `bun build --compile` was turned
down for its binary size. Rust was turned down because of `sjson`. This Go
library sets and deletes a JSON path, and keeps the rest of the file byte for
byte. That is almost all of the `merge` strategy.

Revisit when the plan actions and conflict kinds need more checks than the
`exhaustive` linter gives.

## No copy of `.claude.json`

Claude Code rewrites `.claude.json` every few minutes. The file holds account
details, trust for each project, and caches. A copy on each apply overwrites
that state. A copy made once goes stale and puts account data in git. So Teire
merges owned paths instead.

Revisit when Claude Code keeps user MCP servers in a file that only the user
writes.

## Folder links only on request

A folder link sends each file that a program writes into the folder to the
repository. Claude Code writes transcripts, history and synced skills into its
folder. So a folder link is the `folder` strategy, and a package must ask for
it.

Revisit when most packages are folders that only you write.

## No templates and no fan-out

Templates, and one source file sent to many targets, were turned down. The
personal and work Claude folders share no file, and `merge` covers the
differences inside files seen so far.

Revisit when two targets need the same file, or a file needs a different value
on each machine.

## cobra, not the standard flag package

The standard `flag` package stops at the first argument that is not a flag. It
also makes no help text and no shell completion. cobra gives GNU-style flags,
flags after arguments, help text and zsh completion. It costs two direct
dependencies.

Revisit when those dependencies become a cost.

## Nix with home-manager

By default, home-manager links files into the read-only Nix store. Claude Code
cannot save `/model` or `/config` changes to a read-only `settings.json`.

Revisit when the whole machine moves to Nix.

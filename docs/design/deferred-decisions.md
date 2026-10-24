# Deferred decisions

Questions still open, each with what would settle it. When one is answered,
delete its entry and record the decision in the document it is about.

## Does a running Claude Code session overwrite merged keys?

Claude Code keeps its config in memory and writes `.claude.json` often. It may
write its own copy over the keys that Teire merged. The check after `merge`
reports this, but cannot stop it.

To settle it, merge into `.claude.json` while Claude Code runs. Use Claude Code
for 15 minutes, then run `teire plan`.

## How a new machine gets the binary

A new machine has no Go yet, so `go install` is not a first step there. Release
builds for macOS, attached to each GitHub release, and a Homebrew tap would let
`brew install` or a Brewfile install Teire before it runs. An unsigned binary
may also meet the macOS Gatekeeper, which signing and notarization would avoid.

Settle it before the first release.

# Deferred decisions

Questions still open, each with what would settle it. When one is answered,
delete its entry and record the decision in the document it is about.

## Does a running Claude Code session overwrite merged keys?

Claude Code keeps its config in memory and writes `.claude.json` often. It may
write its own copy over the keys that Teire merged. The check after `merge`
reports this, but cannot stop it.

To settle it, merge into `.claude.json` while Claude Code runs. Use Claude Code
for 15 minutes, then run `teire plan`.

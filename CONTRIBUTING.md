# Contributing

Conventions that apply to work in this repository. Each section covers one kind
of material.

# Setup

```sh
mise trust     # Trust the mise.toml configuration
mise install   # Install Go and golangci-lint
make deps      # Download Go module dependencies
```

# Development workflow

```sh
make check     # Format check, vet and lint
make fix       # Apply formatting and lint fixes
make test      # Run all tests
make build     # Build bin/teire
make install   # Install bin/teire to ~/.local/bin
```

Run `make fix && make check && make test` before pushing.

# Code

Comments explain why, not what or how. If a comment is needed to say what the
code does, rewrite the code instead.

Each package has a package comment in one of its files. `make check` fails
without it.

Put each macOS-only call, such as `scutil` or `defaults`, behind a function
that a test can replace. CI runs on Linux.

## Tests

The tests for `x.go` live in `x_test.go`. Helpers that several test files in
a package share live in `helpers_test.go`, and fixtures live in `testdata/`.

Tests use `testify`. Use `require` for a step that the rest of the test needs,
and `assert` for the checks.

A test that touches files works in `t.TempDir()` and passes that folder as the
home in `Env`. So each test runs against its own temporary home.

# Documentation

Documentation lives in `docs/`: plain Markdown, one subject per file, listed in
the [index][index].

## Organization

Directories partition by domain, such as `architecture/`. They never partition
by document kind, because every document records decisions about its own area
and a folder named for that would describe all of them equally.

A decision and the short reason for it belong in the domain document the
decision is about. Only the alternatives that were turned down, and the
conditions that would justify revisiting them, belong in [tradeoffs][tradeoffs].
Questions still open belong in [deferred decisions][deferred].

A document that lists open questions contains only open questions. When one is
answered, delete the entry. Do not annotate it as resolved, and do not replace
it with a pointer to wherever the answer now lives; either leaves the register
describing something it is not about. The same holds in reverse for any
document scoped to a single kind of content.

Files are named for their subject. Numbered filenames are not used, and neither
is the architecture decision record convention: an ADR is immutable and
superseded rather than edited, which suits an organization keeping a historical
record and does not suit a specification expected to change through
implementation.

## Prose

Documents state what is true. They do not narrate their own history. A rule that
changes is rewritten in place rather than annotated with what it used to be, and
nothing describes a question as newly settled, recently closed, or previously
open. Git holds that record.

Do not use em dashes. Use commas, semicolons, colons, or parentheses.

Wrap lines at 80 columns.

Write headings in sentence case. Use American spelling.

## Links

Use reference-style links rather than inline URLs, so prose stays readable at
80 columns.

Collect the definitions at the end of the document and sort them
alphabetically by key:

```markdown
[deferred]: ../design/deferred-decisions.md
[tradeoffs]: ../design/tradeoffs.md
```

## Tables and diagrams

Diagrams are Mermaid, in fenced blocks.

Pad table columns so the source stays aligned and readable. A table row or a
fenced block that cannot wrap is the only exception to the 80 column rule.

# Git

Branches are named `<type>/<short-description>`, and every change goes through
a pull request; nothing is pushed to `main` directly.

Commit messages follow the conventional commit format:

```text
type(optional scope): subject
```

The type is one of `feat`, `fix`, `chore`, `refactor`, `docs`, or `test`.

Begin the subject with a verb in the imperative mood, so that it completes
the sentence "if applied, this commit will ...". Write `add` rather than
`adds` or `added`, and `fix` rather than `fixed`. Do not end the subject with
a period.

Cap the subject line at 50 characters. Wrap the body at 72 columns. Both are
tighter than the 80 column rule for Markdown, because git indents commit
messages when it displays them.

Close every message with the co-author trailer, exactly as written here and
with no model name:

```text
Co-Authored-By: Claude <noreply@anthropic.com>
```

[deferred]: docs/design/deferred-decisions.md
[index]: docs/index.md
[tradeoffs]: docs/design/tradeoffs.md

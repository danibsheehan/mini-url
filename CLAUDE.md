# CLAUDE.md

@AGENTS.md

The above is the canonical, tool-agnostic reference (install/configure/run/test, conventions,
constraints, definition of done) — also read by Cursor and any other agent. Everything below
is Claude Code–specific session mechanics.

## Skills

`.claude/skills` is the canonical skills directory — add new repo-specific skills here.
Claude Code auto-discovers and invokes them by task:

- `go-conventions` — this repo's Go package layout, handler/service pattern, error handling,
  and test conventions.
- `doc-writer` — README / GoDoc / inline-comment generation for this repo.

This repo also installs the shared `foundations` plugin from the `dani-foundations`
marketplace (see `.claude/settings.json`) for process/docs skills common across projects, and
for Go testing mechanics not specific to this repo (`go-http-testing`, `go-testing`).

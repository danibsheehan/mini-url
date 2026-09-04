# AGENTS.md

Instructions for any coding agent (Cursor, Claude Code, or otherwise) working in this repo.
Human contributors: see `README.md` instead — this file is written for agents and skips the
narrative tour.

mini-url is a minimal Go URL shortener with SQLite persistence and click tracking.

## Stack

- **Go 1.26.1** (see `go.mod`) — stdlib `net/http` only, no router library.
- **Persistence**: SQLite via `github.com/mattn/go-sqlite3`. This driver uses cgo, so a C
  toolchain is required to build.
- No web framework, no ORM — `internal/db` owns connection + schema (created on startup).

## Install

```bash
go mod download
```

Requires a C compiler (Xcode Command Line Tools on macOS, `build-essential` on
Debian/Ubuntu) since `go-sqlite3` uses cgo.

## Configure

No environment variables or flags. The server address (`:8080`) and SQLite path
(`./urls.db`) are hardcoded in `cmd/mini-url/main.go`. If env-based configuration is ever
added, update this section in the same change.

## Run

```bash
make run   # builds and runs the binary
```

## Test / CI parity

```bash
go vet ./...
go build ./...
go test ./...
```

This is the exact sequence `.github/workflows/ci.yml` runs on push/PR. `make test` runs the
test step alone.

## Conventions

See `.claude/skills/go-conventions/SKILL.md` for package layout, the handler/service
pattern, and this repo's test conventions (hand-rolled interface mocks, table-of-subtests,
`t.Fatalf`-only assertions). For the cross-cutting Go testing mechanics and HTTP/interface
fake-shape guidance those conventions build on, see `foundations:go-testing` and
`foundations:go-http-testing`.

## Constraints — do not

- Don't introduce a router library or ORM without discussion — the project is deliberately
  minimal (stdlib `net/http`, raw `database/sql` + `go-sqlite3`).
- Don't add environment-based configuration without updating this file's Configure section
  in the same change.

## Definition of done

- **Task done**: `go test ./...` for the packages you changed (the suite is small enough
  that running it all is usually just as fast).
- **PR done**: `go vet ./...`, `go build ./...`, `go test ./...` — the exact CI sequence.

---

Skills: `.claude/skills` is this repo's canonical skills directory. This repo also installs
the shared `foundations` plugin from the `dani-foundations` marketplace for skills common
across personal projects — see `CLAUDE.md`.

# mini-url

Small URL shortener service implemented in Go.

Getting started
----------------

Requirements:

- Go 1.20+ (project uses modules)

Build and run locally:

```bash
make build
make run
```

Run tests:

```bash
make test
```

Development notes
-----------------

- App entry: `cmd/mini-url/main.go` (package `main`).
- Handlers live in `internal/handlers`.
- Business logic lives in `internal/services` (see `Shortener` interface).
- DB init and connection in `internal/db`.

To reset local state:

```bash
make clean
```

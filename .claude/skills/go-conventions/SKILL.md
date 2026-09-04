---
name: go-conventions
description: >
  Encodes mini-url's own Go conventions for handlers, services, error handling,
  and tests, distilled from the existing codebase. Use this whenever writing or
  reviewing Go code in this repo — new HTTP handlers, service methods, DB logic,
  or their tests — so new code matches existing patterns instead of generic Go
  style. Trigger for requests like "add a new endpoint", "add a service method",
  "write a handler for X", or "add tests for X" within this repo.
---

# mini-url Go Conventions

Patterns observed in `internal/handlers`, `internal/services`, `internal/db`,
and their tests. Follow these instead of inventing new shapes.

## Package layout

- `cmd/mini-url`: entrypoint only. Wires `db.Init` → `services.NewSQLiteShortener`
  → `handlers.NewShortenerHandler` → `http.HandleFunc` routes. No logic here.
- `internal/handlers`: HTTP layer. Depends on a `services.Shortener` interface,
  never on the concrete `sqliteShortener`.
- `internal/services`: business logic. Defines the interface (`Shortener`) and
  its errors in one file (`shortener.go`); the concrete SQLite implementation
  lives in a separate file (`sqlite_shortener.go`).
- `internal/db`: connection + schema management only (`db.Conn`, `db.Init`).

New functionality follows the same split: interface + errors in the services
package, concrete implementation in its own file, handler depends only on the
interface.

## Handler pattern

Every handler method follows this order — copy it for new endpoints:

```go
func (h *ShortenerHandler) Thing(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
        return
    }

    // decode/validate request -> 400 on failure
    // call h.svc.Method(r.Context(), ...)
    // typed/sentinel not-found error -> 404
    // any other error -> generic 500 message, never the raw err
}
```

Rules:
- Method check first, before touching the body.
- `http.Error(w, "<short generic message>", <status>)` — never echo the raw
  `err` back to the client (see `Shorten`'s `"Could not save URL"` vs the
  actual underlying error).
- Not-found errors are checked both ways, since services return either form:
  `if _, ok := err.(*services.NotFoundError); ok || err == services.ErrNotFound`
- Success responses: `w.Header().Set("Content-Type", "application/json")` then
  `json.NewEncoder(w).Encode(resp)` — using a dedicated `<Name>Response` struct
  with `json:"snake_case"` tags, not a map.
- Always pass `r.Context()` into service calls.

## Service pattern

- Interface methods take `ctx context.Context` first, return `(value, error)`.
- Define a sentinel *and* a typed error for "not found" in the same file as
  the interface (`ErrNotFound = &NotFoundError{}`), so callers can check
  either way — matches what the handlers already do.
- Wrap errors with context using `fmt.Errorf("<what was happening>: %w", err)`
  — see `db.Init`'s `"open db: %w"`, `"migrate: %w"` style. Keep the prefix a
  short present-tense phrase describing the failed step.
- SQLite-specific errors (busy/locked/constraint) are classified with small
  named predicate functions (`sqliteBusyOrLocked`, `sqliteUniqueCodeViolation`)
  using `errors.As` against `sqlite3.Error` — don't string-match error text
  outside of `.Error()` equality already established in these helpers.
- Constructors return the interface type, not the concrete struct:
  `func NewSQLiteShortener() Shortener`.

## Tests

- One test file per source file, table-of-subtests via `t.Run("<lowercase
  behavior description>", func(t *testing.T) { ... })` — no `t.Parallel()`,
  no external table slices; each subtest is self-contained.
- Handler tests mock `services.Shortener` with a hand-rolled struct of function
  fields (`mockShortener{createFn, resolveFn, statsFn}`), not a mocking
  library. Every subtest constructs the full mock even if only one field is
  used, wiring unused fields to trivial stubs. See `foundations:go-http-testing`
  for the general guidance on choosing this fake shape (mock the interface)
  over an `httptest.Server` fake (fake the upstream) based on what a handler
  actually depends on.
- Assertions are plain `if got != want { t.Fatalf("<field> = %v, want %v", got, want) }`
  — no assertion library, no `t.Errorf` (use `t.Fatalf` so a subtest stops at
  the first failure).
- Cover: wrong HTTP method, invalid input, not-found path, generic
  service-error path, and the success path — that's the shape used for every
  existing handler method.
- `sqlite_shortener_test.go`'s `TestSQLiteShortener_CreateConcurrentSameOriginalReturnsSameCode`
  proves `Create` is idempotent under concurrent callers (N goroutines racing the same
  original URL, asserting one winning code and one row) — see `foundations:go-testing`'s
  concurrent idempotency/coalescing convergence section for the general shape.
- Retry/backoff around SQLite `busy`/`locked`/constraint errors is tested via an injectable
  code generator (`newSQLiteShortenerWithGenerator`) to force a deterministic code
  collision, plus table-driven `busyRetryBackoff` and `sleepWithContext` cancellation tests
  — see `foundations:go-testing`'s retry/backoff section (this is in fact the pattern that
  section was generalized from).

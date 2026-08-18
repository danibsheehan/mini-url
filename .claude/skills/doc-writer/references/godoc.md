# GoDoc Writing Guide

## Core Rules (from Go conventions)

1. **Every exported symbol must have a doc comment** — `go vet` and `golint` will flag missing ones
2. **Comment goes directly above the declaration** — no blank line between comment and `func`/`type`/`var`
3. **First sentence is the summary** — it appears in package listings. Start with the symbol name.
4. **No `@param` tags** — Go uses prose, not structured tags like JSDoc

## Function Comments

```go
// GetUser retrieves a user by their unique ID.
//
// It queries the primary database and falls back to cache on timeout.
// Returns ErrNotFound if no user with the given ID exists.
// Returns ErrInvalidID if id is not a valid UUID format.
func GetUser(ctx context.Context, id string) (*User, error) {
```

- First line: `// FunctionName verb phrase.`
- Blank comment line = paragraph break in rendered docs
- Describe params in prose only if their purpose isn't obvious from the name + type
- Always document error return values and conditions

## Type Comments

```go
// UserService handles all user lifecycle operations including
// creation, retrieval, and deletion. It is safe for concurrent use.
type UserService struct {
    // db is the backing store. Must not be nil.
    db Database
    // cache is optional; set to nil to disable caching.
    cache Cache
}
```

- Document the struct itself above the type declaration
- Document exported fields inline with `//` comment above each field
- Note concurrency safety if relevant

## Interface Comments

```go
// Store defines the persistence layer for user data.
// Implementations must be safe for concurrent use.
type Store interface {
    // Get retrieves a user by ID. Returns ErrNotFound if absent.
    Get(ctx context.Context, id string) (*User, error)

    // Save persists a user, creating or updating as needed.
    Save(ctx context.Context, u *User) error
}
```

- Document each method in the interface with its own comment
- State error conditions per method

## Package Comments

```go
// Package users provides types and functions for managing user accounts.
//
// Basic usage:
//
//   svc := users.NewService(db)
//   user, err := svc.Get(ctx, "user-id")
//
package users
```

- Always include a package comment in one file (usually `doc.go` or the main file)
- Include a usage example if the package API isn't trivially obvious

## Style Rules

- Use full sentences ending with a period
- Avoid redundancy: don't say "GetUser gets a user" — say what it *does*
- For error-returning functions, always document what errors can be returned
- For context-taking functions, note if cancellation is respected
- Deprecated: `// Deprecated: Use NewFoo instead.` on its own line

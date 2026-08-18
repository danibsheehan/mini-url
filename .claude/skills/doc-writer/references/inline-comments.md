# Inline Code Comments Guide

Inline comments explain *why* code does something — not *what* it does.
The code itself shows *what*. Comments add the context the reader can't get from the code alone.

## The Golden Rule

**Bad** (restates the code):
```ts
// increment i
i++

// check if user is admin
if (user.role === 'admin') {
```

**Good** (explains why or adds non-obvious context):
```ts
// retry index starts at 1; 0 is the initial attempt tracked separately
i++

// only admins can view deleted records per compliance policy
if (user.role === 'admin') {
```

## When to Add Inline Comments

Add a comment when the code:
- Implements a non-obvious algorithm or formula
- Works around a bug in a dependency (cite the issue URL if possible)
- Has a surprising edge case or constraint
- Uses a magic number or hardcoded value
- Does something that looks wrong but is intentional
- Has important performance implications
- Relies on ordering that isn't enforced by the type system

## Format by Language

### TypeScript / JavaScript
```ts
// Single-line: space after //
const TAX_RATE = 0.0875 // CA sales tax as of 2024

/*
 * Multi-line block for longer explanations.
 * Wrap at ~80 chars.
 */
```

### Go
```go
// Single-line: space after //
const maxRetries = 3 // matches upstream service SLA

// TODO(username): remove after migrating to v2 API
```

## Comment Placement

- **Above the line**: for comments explaining a block or statement
- **End of line**: for brief clarifications (keep under ~60 chars)
- Never split a comment across a statement

## What NOT to Comment

- Obvious operations (`// add to slice`, `// return error`)
- Type information already in the signature
- Code that should just be renamed or refactored instead
- Commented-out code (delete it; use version control)

## Density

A well-commented file has comments on maybe 10–20% of lines. More than that usually means the code itself needs to be clearer. When adding inline comments to a file, be selective — focus on the complex, surprising, or business-critical sections.

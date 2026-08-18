# JSDoc Writing Guide (JavaScript / TypeScript)

## When to Use JSDoc vs TypeScript Types

- In **TypeScript** projects: JSDoc is supplementary — types are already in the signature. Focus on `@param` descriptions and `@returns` semantics, not type annotations.
- In **JavaScript** projects: Include `@param {type}` and `@returns {type}` annotations since there's no TypeScript to infer from.

## Standard Function Doc Block

```ts
/**
 * Brief one-line summary. End with a period.
 *
 * Optional longer description if the behavior isn't obvious.
 * Explain side effects, error conditions, async behavior.
 *
 * @param userId - The ID of the user to fetch. Must be a valid UUID.
 * @param options - Optional configuration for the request.
 * @param options.timeout - Request timeout in milliseconds. Defaults to 5000.
 * @returns The user object, or null if not found.
 * @throws {NotFoundError} If the user ID is malformed.
 *
 * @example
 * const user = await getUser('abc-123', { timeout: 3000 })
 * console.log(user.name)
 */
async function getUser(userId: string, options?: GetUserOptions): Promise<User | null>
```

## What to Document

| Symbol | Document? |
|---|---|
| Exported functions | Always |
| Exported classes | Always (class + constructor + public methods) |
| Exported interfaces/types | Always |
| Private/internal functions | Only if complex |
| Simple getters/setters | Only if non-obvious |
| Re-exports | No |

## Tag Reference

- `@param name - description` — describe each parameter's purpose, constraints, defaults
- `@returns description` — what the return value means (not just its type)
- `@throws {ErrorType} condition` — when this throws and with what
- `@example` — short, runnable example (especially for utilities)
- `@deprecated reason` — if replaced, link to the replacement
- `@since version` — optional, for versioned public APIs
- `@internal` — marks symbols not part of the public API

## Style Rules

- First line: imperative verb. "Fetch", "Parse", "Validate" — not "Fetches", not "This function fetches"
- Keep first line under 80 chars
- Describe *why* / *what*, not *how* — callers don't need to know implementation details
- Note gotchas: "Note: mutates the input array", "Throws if called before init()"
- For async functions: mention if it can reject and under what conditions

# mini-url
> A minimal Go URL shortener with SQLite persistence and click tracking.

## Overview
`mini-url` exposes a small HTTP API to create short URLs, redirect short codes to their original targets, and fetch per-link stats. It stores data in SQLite and automatically creates the required schema on startup. The service is intended as a compact learning project and a practical base for extending a URL shortener backend.

## Features
- Creates 6-character short codes for submitted URLs.
- Returns the same short code when a URL is already stored.
- Redirects `/<code>` requests with HTTP 302.
- Tracks redirect click counts for each code.
- Exposes `GET /stats/{code}` with JSON stats output.
- Persists data in a local SQLite database (`urls.db`).
- Includes unit tests for handlers, service logic, and DB initialization.

## Installation
```bash
git clone git@github.com:danibsheehan/mini-url.git
cd mini-url
go mod download
```

## Quick Start
Run the service:

```bash
make run
```

Create a short URL:

```bash
curl -sS -X POST "http://localhost:8080/shorten" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/docs"}'
```

Example response:

```json
{"short_url":"http://localhost:8080/Ab3xYz"}
```

Use the short URL (this increments click count):

```bash
curl -i "http://localhost:8080/Ab3xYz"
```

Fetch stats:

```bash
curl -sS "http://localhost:8080/stats/Ab3xYz"
```

Example response:

```json
{"code":"Ab3xYz","original_url":"https://example.com/docs","click_count":1}
```

## API Endpoints
Base URL (local): `http://localhost:8080`

### `POST /shorten`
Create (or fetch existing) short code for a URL.

Request body:

```json
{"url":"https://example.com"}
```

Success response (`200 OK`):

```json
{"short_url":"http://localhost:8080/Ab3xYz"}
```

Error responses:
- `400 Bad Request` when JSON body is invalid.
- `405 Method Not Allowed` when method is not `POST`.
- `500 Internal Server Error` when persistence fails.

### `GET /{code}`
Redirect to the original URL for a short code.

Success response:
- `302 Found` with `Location: <original_url>`

Error responses:
- `404 Not Found` when code does not exist.
- `500 Internal Server Error` on server-side lookup errors.

Notes:
- `GET /` returns plain text welcome message: `Welcome to Mini URL`.

### `GET /stats/{code}`
Return stats for a short code.

Success response (`200 OK`):

```json
{"code":"Ab3xYz","original_url":"https://example.com","click_count":3}
```

Error responses:
- `404 Not Found` when code does not exist or path is malformed.
- `405 Method Not Allowed` when method is not `GET`.
- `500 Internal Server Error` on server-side lookup errors.

## Curl Cookbook
Copy/paste commands for quick manual testing. This flow uses `jq` to extract the generated code automatically.

### End-to-end happy path
Create a short URL, extract `code`, follow redirect headers, then fetch stats:

```bash
BASE_URL="http://localhost:8080"
TARGET_URL="https://golang.org"

SHORT_URL=$(curl -sS -X POST "$BASE_URL/shorten" \
  -H "Content-Type: application/json" \
  -d "{\"url\":\"$TARGET_URL\"}" | jq -r '.short_url')

CODE="${SHORT_URL##*/}"

echo "SHORT_URL=$SHORT_URL"
echo "CODE=$CODE"

curl -i "$BASE_URL/$CODE"
curl -i "$BASE_URL/stats/$CODE"
```

Expected:
- `POST /shorten` returns `200 OK` with `short_url`.
- `GET /{code}` returns `302 Found` with `Location` header.
- `GET /stats/{code}` returns `200 OK` and JSON with `click_count`.

### Failure examples
Invalid JSON on `POST /shorten`:

```bash
curl -i -X POST "http://localhost:8080/shorten" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://golang.org"'
```

Expected:
- `HTTP/1.1 400 Bad Request`

Unknown code on `GET /{code}`:

```bash
curl -i "http://localhost:8080/notreal"
```

Expected:
- `HTTP/1.1 404 Not Found`

Wrong method on `GET /stats/{code}`:

```bash
curl -i -X POST "http://localhost:8080/stats/notreal"
```

Expected:
- `HTTP/1.1 405 Method Not Allowed`

## Configuration
Current runtime defaults are defined in code:

| Option | Type | Default | Description |
|---|---|---|---|
| Server address | string | `:8080` | HTTP listen address. |
| SQLite path | string | `./urls.db` | Database file used by the service. |

## Development
```bash
make build   # build binary
make run     # build and run locally
make test    # run all tests
make clean   # remove binary and local database file
```

Project layout:
- `cmd/mini-url`: application entrypoint and route wiring.
- `internal/handlers`: HTTP handlers and request/response shaping.
- `internal/services`: shortener business logic and SQLite-backed implementation.
- `internal/db`: database initialization and schema management.

## Contributing
No formal contributing guide is included yet. Open an issue or pull request to propose changes.

## License
MIT. See [LICENSE](./LICENSE).

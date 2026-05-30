# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```sh
go build ./...          # build all packages
go build ./cmd/lrclib   # build the binary
go test ./...           # run all tests
go fmt ./...            # format code
go vet ./...            # static analysis
```

No Makefile, CI config, or external lint config exists yet.

## Architecture

`lrclib` is a Go CLI tool that retrieves synchronized and plain-text lyrics from the [LRCLib.net](https://lrclib.net) API. Module: `github.com/ak1m1tsu/lrclib`, Go 1.26.3, stdlib only.

### Package map

| Package | Path | Role |
|---|---|---|
| `main` | `cmd/lrclib/main.go` | Binary entry point — flag parsing, signal handling, composition root |
| `lrclib` (app) | `internal/app/lrclib/app.go` | `App` struct with `New()` / `Run()` lifecycle |
| `lrclib` (client) | `internal/pkg/lrclib/` | HTTP client for the LRCLib API |

### HTTP client (`internal/pkg/lrclib`)

- `New()` returns a `*Client` with base URL `https://lrclib.net` and 10 s timeout.
- `Client.Get(ctx, *GetRequest) (*GetResponse, error)` — validates the request (track name required), calls `GET /api/get`, decodes JSON.
- Sentinel errors: `ErrTrackNameRequired`, `ErrNotFound`. Unexpected status codes return a formatted error with the status text.
- `GetRequest` fields: `TrackName` (required), `ArtistName`, `AlbumName`, `Duration` (all optional).
- `GetResponse` fields: `ID`, `TrackName`, `ArtistName`, `AlbumName`, `Duration`, `Instrumental`, `PlainLyrics`, `SyncedLyrics`.

### App (`internal/app/lrclib`)

- `Fetcher` interface — defined here (consumer-owned), not in the client package.
- `New(fetcher Fetcher, out io.Writer, opts Options) *App` — wires dependencies; no I/O.
- `App.Run(ctx context.Context) error` — fetches lyrics, selects the appropriate format, writes to `out`. No `os.Exit` inside.
- `selectLyrics(resp, plainOnly) (string, error)` — unexported; prefers synced lyrics, falls back to plain. Returns `ErrTrackIsInstrumental`, `ErrNoPlainLyrics`, or `ErrNoLyrics` on failure.
- `Options` fields: `TrackName`, `ArtistName`, `AlbumName`, `Duration` (int), `PlainOnly` (bool).

### CLI (`cmd/lrclib`)

Track name is a positional argument (first non-flag arg). Flags must come before the positional argument — flags after the track name are silently ignored by `flag.FlagSet`.

```
lrclib [flags] <track name>
```

| Flag | Type | Default | Description |
|---|---|---|---|
| `-artist` | string | `""` | Artist name for search |
| `-album` | string | `""` | Album name for search |
| `-duration` | int | `0` | Track duration in seconds |
| `-plain` | bool | `false` | Output plain lyrics instead of synced LRC |
| `-o` | string | `""` | Write lyrics to file (default: stdout) |

**Exit codes:**

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | Runtime or API error |
| 2 | Usage error (missing track name, bad flags) |

**Example:**

```sh
lrclib -artist "Rick Astley" "Never Gonna Give You Up"
lrclib -artist "Rick Astley" -plain "Never Gonna Give You Up"
lrclib -artist "Rick Astley" -o lyrics.lrc "Never Gonna Give You Up"
```

### Conventions

- All API methods accept a `context.Context` as the first argument.
- Errors are returned as the last return value; `nil` on success.
- `cmd/lrclib` is the only composition root — dependency wiring happens there, not inside `internal/`.

## Web Application

`cmd/web` is a second binary in the same module. It serves a browser UI backed by the same lrclib.net API. Module: `github.com/ak1m1tsu/lrclib`, Go 1.26.3.

### Routes

| Method | Pattern | Handler | Description |
|---|---|---|---|
| GET | `/` | `HomeHandler` | Landing page with search form |
| GET | `/search?q=` | `SearchHandler` | Search results; returns partial for HTMX, full page otherwise |
| GET | `/lyrics/{id}?plain=` | `LyricsHandler` | Lyrics view; returns partial for HTMX, full page otherwise |
| GET | `/static/` | `http.FileServer` | Embedded static assets (htmx.min.js) |

### Middleware chain

Requests pass through `SecureHeaders → Logger → Recover → router` (outermost first).

- `SecureHeaders` — sets `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`, `Referrer-Policy: strict-origin-when-cross-origin`.
- `Logger` — logs method, path, status code, and duration to stderr.
- `Recover` — converts any handler panic into a 500 response.

### Template system

Templates live in `internal/template/` as `.templ` files (a-h/templ DSL). Run `go generate ./internal/template/...` to compile them into `_templ.go` files. Handlers consume templates through narrow renderer interfaces defined in each handler file; `handler.Templates` adapts the generated components to those interfaces.

### HTMX patterns

- **Search-as-you-type** — the search input fires `GET /search?q=` with a 300 ms debounce (`hx-trigger="input delay:300ms, search"`). The response is swapped into `#results`.
- **Lyrics navigation** — track cards send `GET /lyrics/{id}` into `#content` with `hx-push-url="true"`, updating the browser URL without a full page load.
- **Synced/Plain toggle** — the toggle buttons on the lyrics page send `GET /lyrics/{id}?plain=true|false` into `#content` with `hx-push-url="true"`.
- **Partial vs full page** — handlers detect `HX-Request: true` to decide whether to render a component partial or the full layout shell. This makes every route directly bookmarkable.

### Security controls

- Query length is capped at 500 characters in `SearchHandler` before the string reaches the API client.
- Server timeouts: `ReadHeaderTimeout: 5s`, `ReadTimeout: 10s`, `WriteTimeout: 30s`, `IdleTimeout: 60s`.
- `SecureHeaders` middleware applied to all responses.

### Commands (web)

```sh
go generate ./internal/template/...   # recompile Templ files after editing .templ sources
go build -o bin/web ./cmd/web         # build the web binary
go test ./...                         # run all tests
```

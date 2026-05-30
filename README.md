# lrclib

A web app that searches and displays synchronized and plain-text lyrics sourced from [lrclib.net](https://lrclib.net).

## Stack

- Go 1.26+
- [gorilla/mux](https://github.com/gorilla/mux) — HTTP router
- [a-h/templ](https://github.com/a-h/templ) — type-safe HTML templates compiled to Go
- [HTMX](https://htmx.org) — partial page updates without writing JavaScript
- TailwindCSS (CDN)

## Prerequisites

- Go 1.26 or later
- templ CLI: `go install github.com/a-h/templ/cmd/templ@latest`

## Build and run

```sh
go generate ./internal/template/...   # compile .templ files to _templ.go
go build -o bin/web ./cmd/web
./bin/web              # listens on :8080
./bin/web -addr :3000  # custom port
```

## Usage

Open `http://localhost:8080` in a browser. Type a track name or artist into the search box — results appear as you type (300 ms debounce). Click a result to view its lyrics. Use the **Synced** / **Plain** toggle on the lyrics page to switch between the LRC timestamped view and plain text.

## Project layout

| Path | Purpose |
|---|---|
| `cmd/web/` | Binary entry point — flag parsing, routing, server lifecycle |
| `internal/lrclib/` | HTTP client for the lrclib.net API |
| `internal/handler/` | HTTP handlers and middleware (SecureHeaders, Logger, Recover) |
| `internal/template/` | Templ component definitions and generated Go code |
| `cmd/web/static/` | Embedded static assets (htmx.min.js) |

## API source

Lyrics data is fetched from [lrclib.net](https://lrclib.net) — a free, open lyrics API. No API key is required.

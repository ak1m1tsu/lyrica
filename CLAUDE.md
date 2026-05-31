# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```sh
wails build               # build the desktop binary (embeds frontend/dist)
wails dev                 # run in dev mode with hot reload
go test ./...             # run all tests
go fmt ./...              # format code
go vet ./...              # static analysis
```

Frontend only:
```sh
cd frontend && npm run build   # compile React app to frontend/dist
cd frontend && npm run dev     # Vite dev server (used by wails dev)
```

No Makefile, CI config, or external lint config exists yet.

## Architecture

`lrclib` is a Wails v2 desktop application that retrieves synchronized and plain-text lyrics from the [LRCLib.net](https://lrclib.net) API. Module: `github.com/ak1m1tsu/lrclib`, Go 1.26.3.

### Package map

| Package / Path | Role |
|---|---|
| `main.go` | Wails entry point — embeds `frontend/dist`, configures window, binds `App` |
| `app.go` | `App` struct — `Search` and `GetByID` RPC methods exposed to the frontend |
| `internal/lrclib/` | HTTP client for the LRCLib.net API |
| `frontend/src/` | React + TypeScript UI |

### HTTP client (`internal/lrclib`)

- `New()` returns a `*Client` with base URL `https://lrclib.net` and 10 s timeout.
- `Client.Search(ctx, query) ([]Track, error)` — calls `GET /api/search?q=`, decodes JSON array.
- `Client.GetByID(ctx, id) (*Track, error)` — calls `GET /api/get/{id}`, decodes JSON. Returns `ErrNotFound` on 404.
- `Track` fields: `ID`, `TrackName`, `ArtistName`, `AlbumName`, `Duration`, `Instrumental`, `PlainLyrics`, `SyncedLyrics`.

### Desktop backend (`app.go`)

- `NewApp()` creates the `App` and initialises the lrclib HTTP client.
- `startup(ctx)` — Wails lifecycle hook; stores context.
- `App.Search(query string) ([]lrclib.Track, error)` — caps query at 500 chars, delegates to client.
- `App.GetByID(id int) (*lrclib.Track, error)` — delegates to client.
- Both methods are exported and bound to the Wails JS bridge automatically.

### Frontend (`frontend/`)

React 18 + TypeScript + Vite + Tailwind CSS. Wails auto-generates TypeScript bindings in `frontend/wailsjs/` from the `App` struct methods.

| Component | Responsibility |
|---|---|
| `App.tsx` | Top-level state: current view, results, selected track, loading/error flags |
| `SearchBar.tsx` | Debounced search input (300 ms) |
| `ResultsList.tsx` + `TrackCard.tsx` | Track grid with synced/plain/instrumental badges |
| `LyricsView.tsx` | Full lyrics display with synced ↔ plain toggle |
| `EmptyState.tsx` / `ErrorBlock.tsx` | Empty and error states |

### Wails bindings

`frontend/wailsjs/go/main/App.d.ts` (auto-generated):
```ts
export function Search(arg1: string): Promise<Array<lrclib.Track>>;
export function GetByID(arg1: number): Promise<lrclib.Track>;
```

Run `wails generate module` to regenerate bindings after changing exported `App` methods.

### Conventions

- All Go API methods accept `context.Context` as the first argument.
- Errors are returned as the last return value; `nil` on success.
- `main.go` / `app.go` is the only composition root — dependency wiring happens there, not inside `internal/`.

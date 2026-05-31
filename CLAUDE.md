# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```sh
wails build               # build the desktop binary (embeds frontend/dist)
wails dev                 # run in dev mode with hot reload
go test ./...             # run all tests
go fmt ./...              # format code
go vet ./...              # static analysis
wails generate module     # regenerate frontend/wailsjs/ after changing exported App methods
```

Frontend only:
```sh
cd frontend && npm run build   # compile React app to frontend/dist
cd frontend && npm run dev     # Vite dev server (used by wails dev)
```

## Architecture

`lrclib` is a Wails v2 desktop application that retrieves synchronized and plain-text lyrics from the [LRCLib.net](https://lrclib.net) API. Module: `github.com/ak1m1tsu/lrclib`, Go 1.26.3.

### Package map

| Package / Path | Role |
|---|---|
| `main.go` | Wails entry point — embeds `frontend/dist`, configures frameless window (1024×768, min 600×500), binds `App` |
| `app.go` | `App` struct — all RPC methods exposed to the frontend |
| `favorites.go` | `favoritesManager` — favorites persistence, config storage, `sanitizeFilename` |
| `internal/lrclib/` | HTTP client for the LRCLib.net API |
| `frontend/src/` | React + TypeScript UI |

### HTTP client (`internal/lrclib`)

- `New()` returns a `*Client` with base URL `https://lrclib.net`, 10 s timeout, and `User-Agent: lrclib-desktop/1.0`.
- `Client.Search(ctx, query) ([]Track, error)` — `GET /api/search?q=`, decodes JSON array.
- `Client.GetByID(ctx, id) (*Track, error)` — `GET /api/get/{id}`, decodes JSON. Returns `ErrNotFound` on 404.
- `ErrNotFound` — sentinel error variable for 404 responses.
- `Track` fields: `ID`, `TrackName`, `ArtistName`, `AlbumName`, `Duration`, `Instrumental`, `PlainLyrics`, `SyncedLyrics`.

### Favorites (`favorites.go`)

- `favoritesManager` — thread-safe (`sync.RWMutex`) store for favorite tracks plus app config.
- Config file: `{os.UserConfigDir()}/lrclib/config.json` — stores `FavoritesDir` setting.
- Favorites file: `{FavoritesDir}/favorites.json` — JSON array of `Track`.
- `newFavoritesManager()` — loads config and favorites from disk on startup; defaults `FavoritesDir` to the config dir.
- Public methods: `getAll()`, `has(id)`, `add(track)`, `remove(id)`, `getDir()`, `setDir(newDir)`.
- `setDir` migrates `favorites.json` from the old directory to the new one.
- `sanitizeFilename(name) string` — replaces `/ \ : * ? " < > |` with `_`, truncates to 100 chars; used by `ExportLyrics`.

### Desktop backend (`app.go`)

`App` struct holds `ctx`, `*lrclib.Client`, and `*favoritesManager`. `NewApp()` wires them together; `startup(ctx)` stores the Wails context.

All exported methods are bound to the Wails JS bridge automatically:

| Method | Signature | Notes |
|---|---|---|
| `Search` | `(query string) ([]lrclib.Track, error)` | Trims whitespace; caps at 500 chars; returns `[]` (not error) on `ErrNotFound` |
| `GetByID` | `(id int) (*lrclib.Track, error)` | Validates `id > 0`; returns user-friendly error strings on `ErrNotFound` / network error |
| `GetFavorites` | `() []lrclib.Track` | Returns a copy of all saved favorites |
| `AddFavorite` | `(track lrclib.Track) error` | Deduplicates by ID; persists to disk |
| `RemoveFavorite` | `(id int) error` | Removes by ID; persists to disk |
| `IsFavorite` | `(id int) bool` | Thread-safe membership check |
| `GetFavoritesDir` | `() string` | Returns current favorites storage path |
| `PickFavoritesDir` | `() (string, error)` | Opens native directory picker; sets and returns the new path |
| `ExportLyrics` | `(trackName, text, ext string) error` | Opens native save dialog; writes lyrics to `.lrc` or `.txt` |

### Frontend (`frontend/`)

React 18 + TypeScript + Vite + Tailwind CSS (class-based dark mode). Wails auto-generates TypeScript bindings in `frontend/wailsjs/` from the `App` struct methods.

#### Components

| Component | Responsibility |
|---|---|
| `App.tsx` | Top-level state: current view (`home`/`lyrics`), results, selected track, loading/error flags |
| `TitleBar.tsx` | Draggable frameless title bar — theme toggle, favorites button, window controls (minimize/maximize/close via Wails runtime) |
| `SearchBar.tsx` | Search form with submit handler |
| `ResultsList.tsx` | Renders TrackCard list or EmptyState |
| `TrackCard.tsx` | Single result card with synced/plain/instrumental badge and duration |
| `LyricsView.tsx` | Full lyrics display — synced ↔ plain toggle, timestamp column, FavoriteButton, ExportMenu |
| `FavoritesPanel.tsx` | Fixed right-side slide panel listing favorites with remove and folder-change controls |
| `FavoriteButton.tsx` | Filled/unfilled heart icon button; `size` prop (`sm`/`md`) |
| `ExportMenu.tsx` | Dropdown to export `.lrc` or `.txt`; calls `ExportLyrics` RPC |
| `ThemeToggle.tsx` | Sun/moon icon button; delegates to `useTheme` |
| `ProgressBar.tsx` | Animated `animate-progress` bar shown during loading |
| `EmptyState.tsx` / `ErrorBlock.tsx` | Empty and error states |

#### Hooks

- `useTheme()` — returns `{ theme, toggle }`. Persists to `localStorage` key `lrclib-theme`; toggles `.dark` class on `<html>`.
- `useFavorites()` — returns `{ favorites, favoritesDir, isFavorite, toggleFavorite, pickDir }`. Wraps `GetFavorites`, `AddFavorite`, `RemoveFavorite`, `IsFavorite`, `GetFavoritesDir`, `PickFavoritesDir` RPCs.

#### Utilities

- `formatting.ts` — `formatDuration(seconds: number): string` — converts seconds to `M:SS`. Used in TrackCard, LyricsView, FavoritesPanel.

### Wails bindings

`frontend/wailsjs/go/main/App.d.ts` (auto-generated — run `wails generate module` to refresh):

```ts
export function Search(arg1: string): Promise<Array<lrclib.Track>>;
export function GetByID(arg1: number): Promise<lrclib.Track>;
export function GetFavorites(): Promise<Array<lrclib.Track>>;
export function AddFavorite(arg1: lrclib.Track): Promise<void>;
export function RemoveFavorite(arg1: number): Promise<void>;
export function IsFavorite(arg1: number): Promise<boolean>;
export function GetFavoritesDir(): Promise<string>;
export function PickFavoritesDir(): Promise<string>;
export function ExportLyrics(arg1: string, arg2: string, arg3: string): Promise<void>;
```

### Conventions

- All Go API methods accept `context.Context` as the first argument (passed from the Wails runtime via `startup`).
- Errors are returned as the last return value; `nil` on success.
- `main.go` / `app.go` is the only composition root — dependency wiring happens there, not inside `internal/`.
- HTTP errors are wrapped with `fmt.Errorf("lrclib: ...")` in the client; `app.go` translates them to user-friendly strings before returning to the frontend.

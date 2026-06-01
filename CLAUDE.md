# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Versioning

After every implementation, bump the version in `wails.json` (and the `appVersion` constant in `app.go`):

| Change type | Bump |
|---|---|
| `chore`, `fix`, `docs`, `refactor`, `style`, `test` | +0.0.1 (patch) |
| `feat` — new feature, minor update | +0.1.0 (minor) |
| Breaking change / major update | +1.0.0 (major) |

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

**Lyrica** (v3.0.0) is a Wails v2 desktop application that retrieves synchronized and plain-text lyrics from the [LRCLib.net](https://lrclib.net) API. Module: `github.com/ak1m1tsu/lyrica`, Go 1.26.3.

The codebase follows a layered DDD structure: **domain** (no external deps) → **service** (orchestration) → **infrastructure** (I/O implementations).

### Package map

| Package / Path | Role |
|---|---|
| `main.go` | Wails entry point — embeds `frontend/dist`, configures frameless window (1024×768, min 600×500), wires services, binds `App` |
| `app.go` | `App` struct — all RPC methods exposed to the frontend; holds `appVersion` constant |
| `tray.go` | System tray setup using `fyne.io/systray` — Show/Hide and Quit menu items |
| `internal/domain/` | Core `Track` type; port interfaces `LyricsClient` and `FavoritesStore` |
| `internal/service/` | `Lyrics` and `Favorites` use-case orchestration |
| `internal/infrastructure/lrclib/` | HTTP client (plain `Client` + `CachingClient` LRU wrapper) |
| `internal/infrastructure/storage/` | SQLite implementation of `FavoritesStore` (`modernc.org/sqlite` + `squirrel`) |
| `frontend/src/` | React + TypeScript UI |

### HTTP client (`internal/infrastructure/lrclib`)

- `New()` returns a `*Client` with base URL `https://lrclib.net`, 10 s timeout, and `User-Agent: lyrica-desktop/1.0`.
- `Client.Search(ctx, query) ([]Track, error)` — `GET /api/search?q=`, decodes JSON array.
- `Client.GetByID(ctx, id) (*Track, error)` — `GET /api/get/{id}`, decodes JSON. Returns `ErrNotFound` on 404.
- `CachingClient` — LRU cache (256 entries, 10-min TTL) wrapping the base client; same interface.
- `ErrNotFound` — sentinel error variable for 404 responses.
- `Track` fields: `ID`, `TrackName`, `ArtistName`, `AlbumName`, `Duration`, `Instrumental`, `PlainLyrics`, `SyncedLyrics`.

### Storage (`internal/infrastructure/storage`)

- SQLite-backed implementation of `domain.FavoritesStore`.
- Config file: `{os.UserConfigDir()}/lyrica/config.json` — stores `FavoritesDir` and `CloseToTray`.
- Database: `{FavoritesDir}/lyrica.db` — SQLite, replaces the old `favorites.json`.
- Loaded on startup; `FavoritesDir` defaults to the config dir.
- `SetDir` migrates `lyrica.db` from the old directory to the new one.
- `sanitizeFilename(name) string` (in `app.go`) — replaces `/ \ : * ? " < > |` with `_`, truncates to 100 chars; used by `ExportLyrics`.

### Desktop backend (`app.go`)

`App` struct holds `ctx`, `*service.Lyrics`, and `*service.Favorites`. `NewApp()` wires them; `startup(ctx)` stores the Wails context and launches the tray goroutine.

All exported methods are bound to the Wails JS bridge automatically:

| Method | Signature | Notes |
|---|---|---|
| `GetVersion` | `() string` | Returns the `appVersion` constant |
| `GetAppIcon` | `() string` | Returns app icon as base64-encoded PNG data URI |
| `Search` | `(query string) ([]domain.Track, error)` | Trims whitespace; caps at 500 chars; returns `[]` (not error) on `ErrNotFound` |
| `GetByID` | `(id int) (*domain.Track, error)` | Validates `id > 0`; returns user-friendly error strings on `ErrNotFound` / network error |
| `GetFavorites` | `() []domain.Track` | Returns a copy of all saved favorites |
| `AddFavorite` | `(track domain.Track) error` | Deduplicates by ID; persists to DB |
| `RemoveFavorite` | `(id int) error` | Removes by ID; persists to DB |
| `IsFavorite` | `(id int) bool` | Thread-safe membership check |
| `GetFavoritesDir` | `() string` | Returns current favorites storage path |
| `PickFavoritesDir` | `() (string, error)` | Opens native directory picker; sets and returns the new path |
| `ExportLyrics` | `(trackName, text, ext string) error` | Opens native save dialog; writes lyrics to `.lrc` or `.txt` |
| `CloseApp` | `()` | Hides window when close-to-tray is enabled; otherwise quits |
| `GetCloseToTray` | `() bool` | Returns the persisted close-to-tray preference |
| `SetCloseToTray` | `(enabled bool) error` | Persists the close-to-tray preference to config |

### Frontend (`frontend/`)

React 18 + TypeScript + Vite + Tailwind CSS (class-based dark mode). Wails auto-generates TypeScript bindings in `frontend/wailsjs/` from the `App` struct methods.

#### Components

| Component | Responsibility |
|---|---|
| `App.tsx` | Top-level state: current view (`home`/`lyrics`), results, selected track, loading/error flags |
| `TitleBar.tsx` | Draggable frameless title bar — theme toggle, favorites button, settings/about buttons, window controls |
| `SearchBar.tsx` | Search form with submit handler |
| `ResultsList.tsx` | Renders TrackCard list or EmptyState |
| `TrackCard.tsx` | Single result card with synced/plain/instrumental badge and duration |
| `LyricsView.tsx` | Full lyrics display — synced ↔ plain toggle, timestamp column, FavoriteButton, ExportMenu |
| `FavoritesPanel.tsx` | Fixed right-side slide panel listing favorites with remove and folder-change controls |
| `FavoriteButton.tsx` | Filled/unfilled heart icon button; `size` prop (`sm`/`md`) |
| `ExportMenu.tsx` | Dropdown to export `.lrc` or `.txt`; calls `ExportLyrics` RPC |
| `ThemeToggle.tsx` | Sun/moon icon button; delegates to `useTheme` |
| `ProgressBar.tsx` | Animated `animate-progress` bar shown during loading |
| `AboutModal.tsx` | Modal showing app version and icon |
| `SettingsModal.tsx` | Modal for toggling close-to-tray preference via `GetCloseToTray`/`SetCloseToTray` RPCs |
| `EmptyState.tsx` / `ErrorBlock.tsx` | Empty and error states |

#### Hooks

- `useTheme()` — returns `{ theme, toggle }`. Persists to `localStorage` key `lyrica-theme`; toggles `.dark` class on `<html>`.
- `useFavorites()` — returns `{ favorites, favoritesDir, isFavorite, toggleFavorite, pickDir, reload }`. Wraps `GetFavorites`, `AddFavorite`, `RemoveFavorite`, `GetFavoritesDir`, `PickFavoritesDir` RPCs; `reload` manually re-fetches from the backend.
- `useSearch()` — returns `{ results, loading, error, search }`. Debounces queries (300 ms), maintains an in-memory LRU cache (50 entries), guards against out-of-order responses.

#### Utilities

- `formatting.ts` — `formatDuration(seconds: number): string` — converts seconds to `M:SS`. Used in TrackCard, LyricsView, FavoritesPanel.

### Wails bindings

`frontend/wailsjs/go/main/App.d.ts` (auto-generated — run `wails generate module` to refresh):

```ts
export function GetVersion(): Promise<string>;
export function GetAppIcon(): Promise<string>;
export function Search(arg1: string): Promise<Array<domain.Track>>;
export function GetByID(arg1: number): Promise<domain.Track>;
export function GetFavorites(): Promise<Array<domain.Track>>;
export function AddFavorite(arg1: domain.Track): Promise<void>;
export function RemoveFavorite(arg1: number): Promise<void>;
export function IsFavorite(arg1: number): Promise<boolean>;
export function GetFavoritesDir(): Promise<string>;
export function PickFavoritesDir(): Promise<string>;
export function ExportLyrics(arg1: string, arg2: string, arg3: string): Promise<void>;
export function CloseApp(): Promise<void>;
export function GetCloseToTray(): Promise<boolean>;
export function SetCloseToTray(arg1: boolean): Promise<void>;
```

### Conventions

- All Go API methods accept `context.Context` as the first argument (passed from the Wails runtime via `startup`).
- Errors are returned as the last return value; `nil` on success.
- `main.go` / `app.go` is the only composition root — dependency wiring happens there, not inside `internal/`.
- Architecture layers: **domain** has no external deps; **service** orchestrates domain ports; **infrastructure** provides I/O implementations.
- HTTP errors are wrapped with `fmt.Errorf("lrclib: ...")` in the client; `app.go` translates them to user-friendly strings before returning to the frontend.

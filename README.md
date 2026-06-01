# Lyrica

A desktop application that searches and displays synchronized and plain-text lyrics sourced from [lrclib.net](https://lrclib.net). Runs on Windows and macOS.

## Features

- Search tracks by name or artist
- Synced (LRC timestamped) and plain-text lyrics views
- Favorites — persist and manage saved tracks
- Export lyrics as `.lrc` or `.txt` via native save dialog
- Dark / light theme toggle

## Download

Pre-built binaries are published on [GitHub Releases](../../releases):

| Platform | File |
|---|---|
| Windows | `lyrica-amd64-installer.exe` (NSIS installer) |
| macOS | `lyrica-macos.zip` (universal `.app`) |

## Stack

- Go 1.26 + [Wails v2](https://wails.io) — desktop runtime
- React 18 + TypeScript + Vite — frontend
- Tailwind CSS — styling with dark-mode support

## Prerequisites

- Go 1.26+
- Node 20+
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

## Build and run

```sh
wails dev      # dev mode with hot reload
wails build    # production binary (embeds frontend/dist)
```

Frontend only:

```sh
cd frontend && npm run dev
cd frontend && npm run build
```

After `wails generate module`, Wails regenerates the TypeScript bindings in `frontend/wailsjs/` from the exported `App` methods.

## Project layout

| Path | Purpose |
|---|---|
| `main.go` | Wails entry point — frameless window config (1024×768), binds `App` |
| `app.go` | `App` struct — all RPC methods exposed to the frontend |
| `internal/infrastructure/storage/` | SQLite-backed favorites store |
| `internal/infrastructure/lrclib/` | HTTP client for the LRCLib.net API |
| `frontend/src/` | React + TypeScript UI components and hooks |
| `frontend/wailsjs/` | Auto-generated TypeScript bindings (do not edit manually) |

## API source

Lyrics data is fetched from [lrclib.net](https://lrclib.net) — a free, open lyrics API. No API key is required.

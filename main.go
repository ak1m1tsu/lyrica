package main

import (
	"embed"
	"log/slog"
	"os"

	infragithub "github.com/ak1m1tsu/lyrica/internal/infrastructure/github"
	infragdrive "github.com/ak1m1tsu/lyrica/internal/infrastructure/googledrive"
	infralrclib "github.com/ak1m1tsu/lyrica/internal/infrastructure/lrclib"
	"github.com/ak1m1tsu/lyrica/internal/infrastructure/storage"
	"github.com/ak1m1tsu/lyrica/internal/service"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

const (
	windowWidth     = 1024
	windowHeight    = 768
	windowMinWidth  = 600
	windowMinHeight = 500
)

// googleClientID and googleClientSecret are injected at build time via ldflags:
//
//	wails build -ldflags "-X main.googleClientID=<id> -X main.googleClientSecret=<secret>"
var (
	googleClientID     string
	googleClientSecret string
)

func main() {
	store, err := storage.NewSQLiteStore("")
	if err != nil {
		slog.Error("failed to open favorites store", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := store.Close(); err != nil {
			slog.Error("favorites store close failed", "error", err)
		}
	}()
	lyrics := service.NewLyrics(infralrclib.NewCachingClient(infralrclib.New()))
	favorites := service.NewFavorites(store)
	githubClient := infragithub.New()
	updater := service.NewUpdater(appVersion, githubClient)
	gdriveClient := infragdrive.New(googleClientID, googleClientSecret)
	syncService := service.NewSync(store, gdriveClient)
	app := NewApp(lyrics, favorites, updater, gdriveClient, syncService)
	err = wails.Run(&options.App{
		Title:             "Lyrica",
		Width:             windowWidth,
		Height:            windowHeight,
		MinWidth:          windowMinWidth,
		MinHeight:         windowMinHeight,
		Frameless:         true,
		HideWindowOnClose: true,
		BackgroundColour:  &options.RGBA{R: 15, G: 17, B: 23, A: 1},
		AssetServer:       &assetserver.Options{Assets: assets},
		OnStartup:         app.startup,
		Bind:              []any{app},
	})
	if err != nil {
		slog.Error("wails run failed", "error", err)
	}
}

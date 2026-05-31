package main

import (
	"embed"
	"log/slog"

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

func main() {
	app := NewApp()
	err := wails.Run(&options.App{
		Title:            "lrclib",
		Width:            windowWidth,
		Height:           windowHeight,
		MinWidth:         windowMinWidth,
		MinHeight:        windowMinHeight,
		Frameless:        true,
		BackgroundColour: &options.RGBA{R: 15, G: 17, B: 23, A: 1},
		AssetServer:      &assetserver.Options{Assets: assets},
		OnStartup:        app.startup,
		Bind:             []any{app},
	})
	if err != nil {
		slog.Error("wails run failed", "error", err)
	}
}

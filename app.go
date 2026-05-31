package main

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/ak1m1tsu/lrclib/internal/lrclib"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx    context.Context
	client *lrclib.Client
	fav    *favoritesManager
}

func NewApp() *App {
	return &App{client: lrclib.New(), fav: newFavoritesManager()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Search(query string) ([]lrclib.Track, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []lrclib.Track{}, nil
	}
	if len(query) > 500 {
		query = query[:500]
	}
	tracks, err := a.client.Search(a.ctx, query)
	if errors.Is(err, lrclib.ErrNotFound) {
		return []lrclib.Track{}, nil
	}
	return tracks, err
}

func (a *App) GetByID(id int) (*lrclib.Track, error) {
	if id <= 0 {
		return nil, errors.New("invalid track ID")
	}
	track, err := a.client.GetByID(a.ctx, id)
	if errors.Is(err, lrclib.ErrNotFound) {
		return nil, errors.New("Track not found.")
	}
	if err != nil {
		return nil, errors.New("Failed to load lyrics. Please try again.")
	}
	return track, nil
}

func (a *App) GetFavorites() []lrclib.Track {
	return a.fav.getAll()
}

func (a *App) AddFavorite(track lrclib.Track) error {
	return a.fav.add(track)
}

func (a *App) RemoveFavorite(id int) error {
	return a.fav.remove(id)
}

func (a *App) IsFavorite(id int) bool {
	return a.fav.has(id)
}

func (a *App) GetFavoritesDir() string {
	return a.fav.getDir()
}

func (a *App) PickFavoritesDir() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose favorites folder",
	})
	if err != nil || dir == "" {
		return "", err
	}
	if err := a.fav.setDir(dir); err != nil {
		return "", err
	}
	return dir, nil
}

func (a *App) ExportLyrics(trackName, text, ext string) error {
	filters := []runtime.FileFilter{{DisplayName: "LRC files (*.lrc)", Pattern: "*.lrc"}}
	if ext == ".txt" {
		filters = []runtime.FileFilter{{DisplayName: "Text files (*.txt)", Pattern: "*.txt"}}
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export lyrics",
		DefaultFilename: sanitizeFilename(trackName) + ext,
		Filters:         filters,
	})
	if err != nil || path == "" {
		return err
	}
	if !strings.HasSuffix(path, ext) {
		path += ext
	}
	return os.WriteFile(path, []byte(text), 0644)
}

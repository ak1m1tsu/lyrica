package main

import (
	"context"
	"errors"

	"github.com/ak1m1tsu/lrclib/internal/lrclib"
)

type App struct {
	ctx    context.Context
	client *lrclib.Client
}

func NewApp() *App {
	return &App{client: lrclib.New()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Search(query string) ([]lrclib.Track, error) {
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
	track, err := a.client.GetByID(a.ctx, id)
	if errors.Is(err, lrclib.ErrNotFound) {
		return nil, errors.New("Track not found.")
	}
	if err != nil {
		return nil, errors.New("Failed to load lyrics. Please try again.")
	}
	return track, nil
}

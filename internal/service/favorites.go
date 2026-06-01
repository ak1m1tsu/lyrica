package service

import (
	"context"

	"github.com/ak1m1tsu/lyrica/internal/domain"
)

// Favorites exposes the favorites use cases on top of a domain.FavoritesStore.
type Favorites struct {
	store domain.FavoritesStore
}

// NewFavorites returns a Favorites service backed by the given store.
func NewFavorites(store domain.FavoritesStore) *Favorites {
	return &Favorites{store: store}
}

// GetAll returns a copy of all saved favorites.
func (s *Favorites) GetAll() []domain.Track {
	return s.store.GetAll()
}

// IsFavorite reports whether a track with the given ID is a favorite.
func (s *Favorites) IsFavorite(id int) bool {
	return s.store.Has(id)
}

// Add stores a track, deduplicating by ID.
func (s *Favorites) Add(ctx context.Context, track domain.Track) error {
	return s.store.Add(ctx, track)
}

// Remove deletes the favorite with the given ID.
func (s *Favorites) Remove(ctx context.Context, id int) error {
	return s.store.Remove(ctx, id)
}

// Dir returns the current favorites storage directory.
func (s *Favorites) Dir() string {
	return s.store.GetDir()
}

// SetDir changes the favorites storage directory, migrating existing data.
func (s *Favorites) SetDir(ctx context.Context, newDir string) error {
	return s.store.SetDir(ctx, newDir)
}

// CloseToTray returns the persisted close-to-tray preference.
func (s *Favorites) CloseToTray() bool {
	return s.store.GetCloseToTray()
}

// SetCloseToTray persists the close-to-tray preference.
func (s *Favorites) SetCloseToTray(ctx context.Context, enabled bool) error {
	return s.store.SetCloseToTray(ctx, enabled)
}

// DiscordPresence returns the persisted Discord Rich Presence preference.
func (s *Favorites) DiscordPresence() bool {
	return s.store.GetDiscordPresence()
}

// SetDiscordPresence persists the Discord Rich Presence preference.
func (s *Favorites) SetDiscordPresence(ctx context.Context, enabled bool) error {
	return s.store.SetDiscordPresence(ctx, enabled)
}

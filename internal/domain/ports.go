package domain

import "context"

// LyricsClient retrieves lyrics and track metadata from a remote source.
// Implementations live in the infrastructure layer.
type LyricsClient interface {
	// Search returns tracks matching the given free-text query.
	Search(ctx context.Context, query string) ([]Track, error)
	// GetByID returns the track with the given identifier, or ErrNotFound.
	GetByID(ctx context.Context, id int) (*Track, error)
}

// FavoritesStore persists the user's favorite tracks and the directory they
// are stored in. Implementations live in the infrastructure layer.
type FavoritesStore interface {
	// GetAll returns a copy of all saved favorites.
	GetAll() []Track
	// Has reports whether a track with the given ID is a favorite.
	Has(id int) bool
	// Add stores a track, deduplicating by ID.
	Add(ctx context.Context, track Track) error
	// Remove deletes the favorite with the given ID.
	Remove(ctx context.Context, id int) error
	// GetDir returns the current favorites storage directory.
	GetDir() string
	// SetDir changes the favorites storage directory, migrating existing data.
	SetDir(ctx context.Context, newDir string) error
	// GetCloseToTray returns the persisted close-to-tray preference.
	GetCloseToTray() bool
	// SetCloseToTray persists the close-to-tray preference.
	SetCloseToTray(ctx context.Context, enabled bool) error
}

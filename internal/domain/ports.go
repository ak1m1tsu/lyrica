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
	// GetDiscordPresence returns the persisted Discord Rich Presence preference.
	GetDiscordPresence() bool
	// SetDiscordPresence persists the Discord Rich Presence preference.
	SetDiscordPresence(ctx context.Context, enabled bool) error
	// GetSpotifyEnabled returns the persisted Spotify integration preference.
	GetSpotifyEnabled() bool
	// SetSpotifyEnabled persists the Spotify integration preference.
	SetSpotifyEnabled(ctx context.Context, enabled bool) error
	// GetSpotifyAccessToken returns the persisted Spotify access token.
	GetSpotifyAccessToken() string
	// SetSpotifyAccessToken persists the Spotify access token.
	SetSpotifyAccessToken(ctx context.Context, token string) error
	// GetSpotifyRefreshToken returns the persisted Spotify refresh token.
	GetSpotifyRefreshToken() string
	// SetSpotifyRefreshToken persists the Spotify refresh token.
	SetSpotifyRefreshToken(ctx context.Context, token string) error
	// GetSpotifyAutoSearch returns whether auto-search on track change is enabled.
	GetSpotifyAutoSearch() bool
	// SetSpotifyAutoSearch persists the auto-search preference.
	SetSpotifyAutoSearch(ctx context.Context, enabled bool) error
	// GoogleDriveEnabled returns whether Google Drive sync is enabled.
	GoogleDriveEnabled() bool
	// SetGoogleDriveEnabled persists the Google Drive sync preference.
	SetGoogleDriveEnabled(ctx context.Context, enabled bool) error
	// GetGoogleAccessToken returns the persisted Google OAuth access token.
	GetGoogleAccessToken() string
	// SetGoogleAccessToken persists the Google OAuth access token.
	SetGoogleAccessToken(ctx context.Context, token string) error
	// GetGoogleRefreshToken returns the persisted Google OAuth refresh token.
	GetGoogleRefreshToken() string
	// SetGoogleRefreshToken persists the Google OAuth refresh token.
	SetGoogleRefreshToken(ctx context.Context, token string) error
	// GetLastSyncAt returns the RFC3339 timestamp of the last successful sync.
	GetLastSyncAt() string
	// SetLastSyncAt persists the last successful sync timestamp.
	SetLastSyncAt(ctx context.Context, t string) error
	// GetCurrentTheme returns the ID of the currently active theme.
	GetCurrentTheme() string
	// SetCurrentTheme persists the active theme ID.
	SetCurrentTheme(ctx context.Context, id string) error
	// GetThemesDir returns the directory where custom theme JSON files are stored.
	GetThemesDir() string
	// GetCustomThemes returns all custom themes found in the themes directory.
	GetCustomThemes() ([]Theme, error)
	// SaveCustomTheme writes a theme to {themesDir}/{id}.json.
	SaveCustomTheme(theme Theme) error
	// DeleteCustomTheme removes {themesDir}/{id}.json.
	DeleteCustomTheme(id string) error
}

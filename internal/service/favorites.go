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

// SpotifyEnabled returns the persisted Spotify integration preference.
func (s *Favorites) SpotifyEnabled() bool {
	return s.store.GetSpotifyEnabled()
}

// SetSpotifyEnabled persists the Spotify integration preference.
func (s *Favorites) SetSpotifyEnabled(ctx context.Context, enabled bool) error {
	return s.store.SetSpotifyEnabled(ctx, enabled)
}

// SpotifyAccessToken returns the persisted Spotify access token.
func (s *Favorites) SpotifyAccessToken() string {
	return s.store.GetSpotifyAccessToken()
}

// SetSpotifyAccessToken persists the Spotify access token.
func (s *Favorites) SetSpotifyAccessToken(ctx context.Context, token string) error {
	return s.store.SetSpotifyAccessToken(ctx, token)
}

// SpotifyRefreshToken returns the persisted Spotify refresh token.
func (s *Favorites) SpotifyRefreshToken() string {
	return s.store.GetSpotifyRefreshToken()
}

// SetSpotifyRefreshToken persists the Spotify refresh token.
func (s *Favorites) SetSpotifyRefreshToken(ctx context.Context, token string) error {
	return s.store.SetSpotifyRefreshToken(ctx, token)
}

// SpotifyAutoSearch returns the persisted auto-search preference.
func (s *Favorites) SpotifyAutoSearch() bool {
	return s.store.GetSpotifyAutoSearch()
}

// SetSpotifyAutoSearch persists the auto-search preference.
func (s *Favorites) SetSpotifyAutoSearch(ctx context.Context, enabled bool) error {
	return s.store.SetSpotifyAutoSearch(ctx, enabled)
}

// GoogleDriveEnabled returns whether Google Drive sync is enabled.
func (s *Favorites) GoogleDriveEnabled() bool {
	return s.store.GoogleDriveEnabled()
}

// SetGoogleDriveEnabled persists the Google Drive sync preference.
func (s *Favorites) SetGoogleDriveEnabled(ctx context.Context, enabled bool) error {
	return s.store.SetGoogleDriveEnabled(ctx, enabled)
}

// GoogleAccessToken returns the persisted Google OAuth access token.
func (s *Favorites) GoogleAccessToken() string {
	return s.store.GetGoogleAccessToken()
}

// SetGoogleAccessToken persists the Google OAuth access token.
func (s *Favorites) SetGoogleAccessToken(ctx context.Context, token string) error {
	return s.store.SetGoogleAccessToken(ctx, token)
}

// GoogleRefreshToken returns the persisted Google OAuth refresh token.
func (s *Favorites) GoogleRefreshToken() string {
	return s.store.GetGoogleRefreshToken()
}

// SetGoogleRefreshToken persists the Google OAuth refresh token.
func (s *Favorites) SetGoogleRefreshToken(ctx context.Context, token string) error {
	return s.store.SetGoogleRefreshToken(ctx, token)
}

// LastSyncAt returns the RFC3339 timestamp of the last successful sync.
func (s *Favorites) LastSyncAt() string {
	return s.store.GetLastSyncAt()
}

// SetLastSyncAt persists the last successful sync timestamp.
func (s *Favorites) SetLastSyncAt(ctx context.Context, t string) error {
	return s.store.SetLastSyncAt(ctx, t)
}

// CurrentTheme returns the ID of the currently active theme.
func (s *Favorites) CurrentTheme() string {
	return s.store.GetCurrentTheme()
}

// SetCurrentTheme persists the active theme ID.
func (s *Favorites) SetCurrentTheme(ctx context.Context, id string) error {
	return s.store.SetCurrentTheme(ctx, id)
}

// ThemesDir returns the directory where custom theme JSON files are stored.
func (s *Favorites) ThemesDir() string {
	return s.store.GetThemesDir()
}

// CustomThemes returns all custom themes from the themes directory.
func (s *Favorites) CustomThemes() ([]domain.Theme, error) {
	return s.store.GetCustomThemes()
}

// SaveCustomTheme writes a theme to the themes directory.
func (s *Favorites) SaveCustomTheme(theme domain.Theme) error {
	return s.store.SaveCustomTheme(theme)
}

// DeleteCustomTheme removes a custom theme by ID from the themes directory.
func (s *Favorites) DeleteCustomTheme(id string) error {
	return s.store.DeleteCustomTheme(id)
}

// AutoUpdateEnabled returns whether automatic background updates are enabled.
func (s *Favorites) AutoUpdateEnabled() bool {
	return s.store.AutoUpdateEnabled()
}

// SetAutoUpdateEnabled persists the auto-update preference.
func (s *Favorites) SetAutoUpdateEnabled(ctx context.Context, enabled bool) error {
	return s.store.SetAutoUpdateEnabled(ctx, enabled)
}

// CheckUpdatesEnabled returns whether the app should check for updates at all.
func (s *Favorites) CheckUpdatesEnabled() bool {
	return s.store.CheckUpdatesEnabled()
}

// SetCheckUpdatesEnabled persists the check-for-updates preference.
func (s *Favorites) SetCheckUpdatesEnabled(ctx context.Context, enabled bool) error {
	return s.store.SetCheckUpdatesEnabled(ctx, enabled)
}

// SkippedVersion returns the version string the user chose to skip.
func (s *Favorites) SkippedVersion() string {
	return s.store.SkippedVersion()
}

// SetSkippedVersion persists the version the user chose to skip.
func (s *Favorites) SetSkippedVersion(ctx context.Context, version string) error {
	return s.store.SetSkippedVersion(ctx, version)
}

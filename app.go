package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ak1m1tsu/lyrica/internal/domain"
	infragithub "github.com/ak1m1tsu/lyrica/internal/infrastructure/github"
	infragdrive "github.com/ak1m1tsu/lyrica/internal/infrastructure/googledrive"
	"github.com/ak1m1tsu/lyrica/internal/service"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const appVersion = "3.8.0"

// UpdateResult is the JSON-serialisable payload returned to the frontend by
// update-related RPC methods and emitted on the "update:available" event.
type UpdateResult struct {
	Available      bool   `json:"available"`
	LatestVersion  string `json:"latestVersion"`
	DownloadURL    string `json:"downloadURL"`
	InstallerName  string `json:"installerName"`
	AssetSizeBytes int64  `json:"assetSizeBytes"`
}

// App is the Wails-bound adapter. It owns the Wails runtime context and
// delegates all business logic to the injected services.
type App struct {
	ctx                context.Context
	lyrics             *service.Lyrics
	favorites          *service.Favorites
	updater            *service.Updater
	discord            *discordPresence
	spotify            *spotifyService
	gdrive             *infragdrive.Client
	sync               *service.Sync
	windowVisible      bool
	updateAvailable    atomic.Bool
	cachedUpdate       atomic.Pointer[service.UpdateInfo]
	downloadInProgress atomic.Bool
}

// NewApp wires the infrastructure into the services and returns the adapter.
func NewApp(lyrics *service.Lyrics, favorites *service.Favorites, updater *service.Updater, gdrive *infragdrive.Client, sync *service.Sync) *App {
	return &App{
		lyrics:    lyrics,
		favorites: favorites,
		updater:   updater,
		discord:   newDiscordPresence(),
		spotify:   newSpotifyService(),
		gdrive:    gdrive,
		sync:      sync,
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.windowVisible = true
	if a.favorites.DiscordPresence() {
		a.discord.connect()
		a.discord.setIdle()
	}
	if a.favorites.SpotifyEnabled() {
		if access := a.favorites.SpotifyAccessToken(); access != "" {
			auth := spotifyauth.New(
				spotifyauth.WithRedirectURL(spotifyRedirectURI),
				spotifyauth.WithScopes(spotifyauth.ScopeUserReadCurrentlyPlaying),
				spotifyauth.WithClientID(spotifyClientID),
			)
			tok := &oauth2.Token{
				AccessToken:  access,
				RefreshToken: a.favorites.SpotifyRefreshToken(),
				TokenType:    "Bearer",
			}
			a.spotify.startPolling(ctx, auth, tok, a.onSpotifyToken, a.onSpotifyTrack, a.onSpotifyError)
		}
	}
	go func() {
		ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
		defer cancel()
		info, err := a.updater.CheckForUpdate(ctx)
		if err != nil || !info.Available {
			return
		}
		a.updateAvailable.Store(true)
		a.cachedUpdate.Store(info)
		runtime.EventsEmit(a.ctx, "update:available", UpdateResult{
			Available:      true,
			LatestVersion:  info.LatestVersion,
			DownloadURL:    info.DownloadURL,
			InstallerName:  info.InstallerName,
			AssetSizeBytes: info.AssetSize,
		})
	}()
	go a.runTray()
}

// GetVersion returns the application version string.
func (a *App) GetVersion() string {
	return appVersion
}

// GetAppIcon returns the app icon as a base64-encoded PNG data URI.
func (a *App) GetAppIcon() string {
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(appIconData)
}

// Search returns tracks matching the query, or an empty slice when none match.
func (a *App) Search(query string) ([]domain.Track, error) {
	return a.lyrics.Search(a.ctx, query)
}

// GetByID returns a single track, translating service errors into
// user-friendly messages for the frontend.
func (a *App) GetByID(id int) (*domain.Track, error) {
	track, err := a.lyrics.GetByID(a.ctx, id)
	switch {
	case errors.Is(err, service.ErrInvalidID):
		return nil, errors.New("Invalid track ID.")
	case errors.Is(err, domain.ErrNotFound):
		return nil, errors.New("Track not found.")
	case err != nil:
		return nil, errors.New("Failed to load lyrics. Please try again.")
	}
	return track, nil
}

// GetFavorites returns a copy of all saved favorites.
func (a *App) GetFavorites() []domain.Track {
	return a.favorites.GetAll()
}

// AddFavorite stores a track, deduplicating by ID.
func (a *App) AddFavorite(track domain.Track) error {
	return a.favorites.Add(a.ctx, track)
}

// RemoveFavorite removes a favorite by ID.
func (a *App) RemoveFavorite(id int) error {
	return a.favorites.Remove(a.ctx, id)
}

// IsFavorite reports whether a track is a favorite.
func (a *App) IsFavorite(id int) bool {
	return a.favorites.IsFavorite(id)
}

// GetFavoritesDir returns the current favorites storage path.
func (a *App) GetFavoritesDir() string {
	return a.favorites.Dir()
}

// PickFavoritesDir opens a native directory picker and persists the choice.
func (a *App) PickFavoritesDir() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Choose favorites folder",
	})
	if err != nil || dir == "" {
		return "", err
	}
	if err := a.favorites.SetDir(a.ctx, dir); err != nil {
		return "", err
	}
	return dir, nil
}

// ExportLyrics opens a native save dialog and writes lyrics to a .lrc or .txt file.
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

// CloseApp hides the window when close-to-tray is enabled, otherwise quits.
func (a *App) CloseApp() {
	if a.favorites.CloseToTray() {
		runtime.WindowHide(a.ctx)
		a.windowVisible = false
		return
	}
	a.discord.disconnect()
	a.spotify.stopPolling()
	runtime.Quit(a.ctx)
}

// GetCloseToTray returns the persisted close-to-tray preference.
func (a *App) GetCloseToTray() bool {
	return a.favorites.CloseToTray()
}

// SetCloseToTray persists the close-to-tray preference.
func (a *App) SetCloseToTray(enabled bool) error {
	return a.favorites.SetCloseToTray(a.ctx, enabled)
}

// GetCurrentTheme returns the ID of the currently active theme.
// Returns "" when no theme has been saved (defaults to "dark" on the frontend).
func (a *App) GetCurrentTheme() string {
	return a.favorites.CurrentTheme()
}

// SetCurrentTheme persists the active theme ID.
func (a *App) SetCurrentTheme(id string) error {
	return a.favorites.SetCurrentTheme(a.ctx, id)
}

// GetCustomThemes returns all user-created themes stored as JSON files.
func (a *App) GetCustomThemes() ([]domain.Theme, error) {
	return a.favorites.CustomThemes()
}

// SaveCustomTheme creates or updates a custom theme JSON file.
// The ID must be non-empty, alphanumeric with hyphens/underscores only, and
// must not conflict with the built-in IDs "light" or "dark".
func (a *App) SaveCustomTheme(theme domain.Theme) error {
	if theme.ID == "" || theme.Name == "" {
		return errors.New("Theme ID and name are required.")
	}
	if theme.ID == "light" || theme.ID == "dark" {
		return errors.New("Cannot overwrite built-in themes.")
	}
	for _, c := range theme.ID {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return errors.New("Theme ID may only contain lowercase letters, digits, hyphens, and underscores.")
		}
	}
	return a.favorites.SaveCustomTheme(theme)
}

// DeleteCustomTheme removes a custom theme by ID.
// Refuses to delete the built-in "light" and "dark" themes.
func (a *App) DeleteCustomTheme(id string) error {
	if id == "light" || id == "dark" {
		return errors.New("Cannot delete built-in themes.")
	}
	return a.favorites.DeleteCustomTheme(id)
}

// ExportTheme opens a native save dialog and writes the custom theme as a JSON file.
func (a *App) ExportTheme(id string) error {
	themes, err := a.favorites.CustomThemes()
	if err != nil {
		return err
	}
	var theme *domain.Theme
	for i := range themes {
		if themes[i].ID == id {
			theme = &themes[i]
			break
		}
	}
	if theme == nil {
		return errors.New("Theme not found.")
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export theme",
		DefaultFilename: sanitizeFilename(theme.Name) + ".json",
		Filters:         []runtime.FileFilter{{DisplayName: "JSON files (*.json)", Pattern: "*.json"}},
	})
	if err != nil || path == "" {
		return err
	}
	if !strings.HasSuffix(path, ".json") {
		path += ".json"
	}
	data, err := json.MarshalIndent(theme, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// ImportTheme opens a native file picker, reads the selected JSON theme,
// saves it to the themes directory, and sets it as the active theme.
// Returns the imported theme so the frontend can update its state immediately.
func (a *App) ImportTheme() (*domain.Theme, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Import theme",
		Filters: []runtime.FileFilter{{DisplayName: "JSON files (*.json)", Pattern: "*.json"}},
	})
	if err != nil || path == "" {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read theme file: %w", err)
	}
	var theme domain.Theme
	if err := json.Unmarshal(data, &theme); err != nil {
		return nil, errors.New("Invalid theme file format.")
	}
	if theme.ID == "" || theme.Name == "" {
		return nil, errors.New("Theme file is missing required id or name fields.")
	}
	if theme.ID == "light" || theme.ID == "dark" {
		return nil, errors.New("Cannot import a theme with a reserved built-in ID.")
	}
	if err := a.favorites.SaveCustomTheme(theme); err != nil {
		return nil, err
	}
	if err := a.favorites.SetCurrentTheme(a.ctx, theme.ID); err != nil {
		return nil, err
	}
	return &theme, nil
}

// GetDiscordPresence returns whether Discord Rich Presence is enabled.
func (a *App) GetDiscordPresence() bool {
	return a.favorites.DiscordPresence()
}

// SetDiscordPresence enables or disables Discord Rich Presence.
func (a *App) SetDiscordPresence(enabled bool) error {
	if err := a.favorites.SetDiscordPresence(a.ctx, enabled); err != nil {
		return err
	}
	if enabled {
		a.discord.connect()
		a.discord.setIdle()
	} else {
		a.discord.disconnect()
	}
	return nil
}

// UpdatePresenceIdle resets Discord presence to the idle browsing state.
func (a *App) UpdatePresenceIdle() {
	a.discord.setIdle()
}

// UpdatePresenceSearching sets Discord presence to reflect an active search query.
func (a *App) UpdatePresenceSearching(query string) {
	a.discord.setSearching(query)
}

// UpdatePresenceTrack sets Discord presence to the currently-viewed track.
func (a *App) UpdatePresenceTrack(trackName, artistName string, synced bool) {
	a.discord.setTrack(trackName, artistName, synced)
}

// onSpotifyToken persists refreshed Spotify tokens after a polling refresh.
func (a *App) onSpotifyToken(access, refresh string) {
	_ = a.favorites.SetSpotifyAccessToken(a.ctx, access)
	if refresh != "" {
		_ = a.favorites.SetSpotifyRefreshToken(a.ctx, refresh)
	}
}

// onSpotifyTrack emits the currently-playing Spotify track to the frontend,
// but only when auto-search is enabled.
func (a *App) onSpotifyTrack(trackName, artistName string) {
	if !a.favorites.SpotifyAutoSearch() {
		return
	}
	runtime.EventsEmit(a.ctx, "spotify:track", map[string]string{
		"trackName":  trackName,
		"artistName": artistName,
	})
}

// onSpotifyError notifies the frontend that the Spotify token has expired.
func (a *App) onSpotifyError(_ error) {
	runtime.EventsEmit(a.ctx, "spotify:token_expired", nil)
}

// GetSpotifyEnabled returns whether the Spotify integration is enabled.
func (a *App) GetSpotifyEnabled() bool {
	return a.favorites.SpotifyEnabled()
}

// GetSpotifyAutoSearch returns whether auto-search on track change is enabled.
func (a *App) GetSpotifyAutoSearch() bool {
	return a.favorites.SpotifyAutoSearch()
}

// SetSpotifyAutoSearch persists the auto-search preference.
func (a *App) SetSpotifyAutoSearch(enabled bool) error {
	return a.favorites.SetSpotifyAutoSearch(a.ctx, enabled)
}

// SetSpotifyEnabled enables or disables the Spotify integration preference.
func (a *App) SetSpotifyEnabled(enabled bool) error {
	return a.favorites.SetSpotifyEnabled(a.ctx, enabled)
}

// ConnectSpotify runs the OAuth PKCE flow on a local HTTPS server, persists
// the tokens, and starts polling.
func (a *App) ConnectSpotify() error {
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(spotifyRedirectURI),
		spotifyauth.WithScopes(spotifyauth.ScopeUserReadCurrentlyPlaying),
		spotifyauth.WithClientID(spotifyClientID),
	)
	verifier := oauth2.GenerateVerifier()
	state := "lyrica-spotify"
	authURL := auth.AuthURL(state, oauth2.S256ChallengeOption(verifier))

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	mux := http.NewServeMux()
	srv := &http.Server{Addr: ":27182", Handler: mux}
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			select {
			case errCh <- errors.New("state mismatch"):
			default:
			}
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			select {
			case errCh <- errors.New("no code in callback"):
			default:
			}
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}
		select {
		case codeCh <- code:
		default:
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h2>Lyrica connected to Spotify!</h2><p>You can close this tab.</p></body></html>")
	})
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case errCh <- err:
			default:
			}
		}
	}()
	runtime.BrowserOpenURL(a.ctx, authURL)

	ctx, cancel := context.WithTimeout(a.ctx, 120*time.Second)
	defer cancel()
	defer srv.Shutdown(context.Background()) //nolint:errcheck

	select {
	case code := <-codeCh:
		token, err := auth.Exchange(ctx, code, oauth2.VerifierOption(verifier))
		if err != nil {
			return fmt.Errorf("Spotify authorization failed: %w", err)
		}
		if err := a.favorites.SetSpotifyAccessToken(a.ctx, token.AccessToken); err != nil {
			return err
		}
		if err := a.favorites.SetSpotifyRefreshToken(a.ctx, token.RefreshToken); err != nil {
			return err
		}
		if err := a.favorites.SetSpotifyEnabled(a.ctx, true); err != nil {
			return err
		}
		a.spotify.stopPolling()
		a.spotify.startPolling(a.ctx, auth, token, a.onSpotifyToken, a.onSpotifyTrack, a.onSpotifyError)
		runtime.EventsEmit(a.ctx, "spotify:connected", nil)
		return nil
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return errors.New("Spotify authorization timed out. Please try again.")
	}
}

// DisconnectSpotify stops polling and clears the persisted Spotify credentials.
func (a *App) DisconnectSpotify() error {
	a.spotify.stopPolling()
	if err := a.favorites.SetSpotifyEnabled(a.ctx, false); err != nil {
		return err
	}
	if err := a.favorites.SetSpotifyAccessToken(a.ctx, ""); err != nil {
		return err
	}
	return a.favorites.SetSpotifyRefreshToken(a.ctx, "")
}

// GetGoogleDriveEnabled returns whether Google Drive sync is enabled.
func (a *App) GetGoogleDriveEnabled() bool {
	return a.favorites.GoogleDriveEnabled()
}

// GetLastSyncTime returns the RFC3339 timestamp of the last successful sync.
func (a *App) GetLastSyncTime() string {
	return a.favorites.LastSyncAt()
}

// ConnectGoogleDrive runs the OAuth2 PKCE flow on a local HTTP server, persists
// the tokens, and enables Google Drive sync.
func (a *App) ConnectGoogleDrive() error {
	if !a.gdrive.Configured() {
		return errors.New("Google Drive credentials are not configured.")
	}
	verifier := oauth2.GenerateVerifier()
	state := "lyrica-gdrive"
	authURL := a.gdrive.AuthURL(state, verifier)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	mux := http.NewServeMux()
	srv := &http.Server{Addr: ":27183", Handler: mux}
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			select {
			case errCh <- errors.New("state mismatch"):
			default:
			}
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			select {
			case errCh <- errors.New("no code in callback"):
			default:
			}
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}
		select {
		case codeCh <- code:
		default:
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><body><h2>Lyrica connected to Google Drive!</h2><p>You can close this tab.</p></body></html>")
	})
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case errCh <- err:
			default:
			}
		}
	}()
	runtime.BrowserOpenURL(a.ctx, authURL)

	ctx, cancel := context.WithTimeout(a.ctx, 120*time.Second)
	defer cancel()
	defer srv.Shutdown(context.Background()) //nolint:errcheck

	select {
	case code := <-codeCh:
		token, err := a.gdrive.Exchange(ctx, code, verifier)
		if err != nil {
			return fmt.Errorf("Google Drive authorization failed: %w", err)
		}
		if err := a.favorites.SetGoogleAccessToken(a.ctx, token.AccessToken); err != nil {
			return err
		}
		if err := a.favorites.SetGoogleRefreshToken(a.ctx, token.RefreshToken); err != nil {
			return err
		}
		return a.favorites.SetGoogleDriveEnabled(a.ctx, true)
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return errors.New("Google Drive authorization timed out. Please try again.")
	}
}

// DisconnectGoogleDrive clears the persisted Google Drive credentials and
// disables sync.
func (a *App) DisconnectGoogleDrive() error {
	if err := a.favorites.SetGoogleDriveEnabled(a.ctx, false); err != nil {
		return err
	}
	if err := a.favorites.SetGoogleAccessToken(a.ctx, ""); err != nil {
		return err
	}
	if err := a.favorites.SetGoogleRefreshToken(a.ctx, ""); err != nil {
		return err
	}
	return a.favorites.SetLastSyncAt(a.ctx, "")
}

// SyncToGoogleDrive uploads the local favorites list to Google Drive.
func (a *App) SyncToGoogleDrive() error {
	if err := a.sync.Push(a.ctx); err != nil {
		if infragdrive.IsAuthError(err) {
			return errors.New("token_expired")
		}
		return fmt.Errorf("Upload failed: %w", err)
	}
	runtime.EventsEmit(a.ctx, "gdrive:synced", a.favorites.LastSyncAt())
	return nil
}

// SyncFromGoogleDrive downloads favorites from Google Drive and merges them
// into the local store, emitting "gdrive:synced" on success.
func (a *App) SyncFromGoogleDrive() error {
	if _, err := a.sync.Pull(a.ctx); err != nil {
		if infragdrive.IsAuthError(err) {
			return errors.New("token_expired")
		}
		return fmt.Errorf("Download failed: %w", err)
	}
	runtime.EventsEmit(a.ctx, "gdrive:synced", a.favorites.LastSyncAt())
	return nil
}

// CheckForUpdates checks GitHub Releases for a newer version and returns the
// result. It also caches the updateAvailable flag for GetUpdateAvailable.
func (a *App) CheckForUpdates() (*UpdateResult, error) {
	ctx, cancel := context.WithTimeout(a.ctx, 30*time.Second)
	defer cancel()
	info, err := a.updater.CheckForUpdate(ctx)
	if err != nil {
		return nil, errors.New("Failed to check for updates. Please try again.")
	}
	a.updateAvailable.Store(info.Available)
	if info.Available {
		a.cachedUpdate.Store(info)
	}
	return &UpdateResult{
		Available:      info.Available,
		LatestVersion:  info.LatestVersion,
		DownloadURL:    info.DownloadURL,
		InstallerName:  info.InstallerName,
		AssetSizeBytes: info.AssetSize,
	}, nil
}

// GetUpdateAvailable returns whether a newer version is known to exist based
// on the most recent check (startup or manual).
func (a *App) GetUpdateAvailable() bool {
	return a.updateAvailable.Load()
}

// DownloadAndInstall downloads the latest installer and launches it, then
// quits the app. It is guarded by an atomic flag to prevent concurrent calls.
func (a *App) DownloadAndInstall() error {
	if !a.downloadInProgress.CompareAndSwap(false, true) {
		return errors.New("A download is already in progress.")
	}
	defer a.downloadInProgress.Store(false)

	info := a.cachedUpdate.Load()
	if info == nil || !info.Available {
		return errors.New("No update available.")
	}

	dlCtx, dlCancel := context.WithTimeout(a.ctx, 10*time.Minute)
	defer dlCancel()
	var lastEmit time.Time
	installerPath, err := infragithub.DownloadInstaller(dlCtx, info.DownloadURL, func(received, total int64) {
		now := time.Now()
		if now.Sub(lastEmit) < 250*time.Millisecond && !(total > 0 && received == total) {
			return
		}
		lastEmit = now
		runtime.EventsEmit(a.ctx, "update:progress", map[string]any{
			"received": received,
			"total":    total,
		})
	})
	if err != nil {
		return errors.New("Download failed. Please try again.")
	}

	if err := a.launchInstallerAndQuit(installerPath); err != nil {
		return errors.New("Failed to launch installer. Please run it manually.")
	}
	return nil
}


// sanitizeFilename replaces filesystem-reserved characters with underscores
// and truncates the result to 100 characters.
func sanitizeFilename(name string) string {
	r := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", `"`, "_", "<", "_", ">", "_", "|", "_",
	)
	result := r.Replace(name)
	if len(result) > 100 {
		result = result[:100]
	}
	return result
}

package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/ak1m1tsu/lyrica/internal/domain"
	infragithub "github.com/ak1m1tsu/lyrica/internal/infrastructure/github"
	"github.com/ak1m1tsu/lyrica/internal/service"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const appVersion = "3.4.0"

// UpdateResult is the JSON-serialisable payload returned to the frontend by
// update-related RPC methods and emitted on the "update:available" event.
type UpdateResult struct {
	Available      bool   `json:"available"`
	LatestVersion  string `json:"latestVersion"`
	ReleaseNotes   string `json:"releaseNotes"`
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
	windowVisible      bool
	updateAvailable    atomic.Bool
	cachedUpdate       atomic.Pointer[service.UpdateInfo]
	downloadInProgress atomic.Bool
}

// NewApp wires the infrastructure into the services and returns the adapter.
func NewApp(lyrics *service.Lyrics, favorites *service.Favorites, updater *service.Updater) *App {
	return &App{
		lyrics:    lyrics,
		favorites: favorites,
		updater:   updater,
		discord:   newDiscordPresence(),
		spotify:   newSpotifyService(),
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
			a.spotify.startPolling(ctx, auth, tok, a.onSpotifyToken, a.onSpotifyTrack)
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
			ReleaseNotes:   info.ReleaseNotes,
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

// onSpotifyTrack emits the currently-playing Spotify track to the frontend.
func (a *App) onSpotifyTrack(trackName, artistName string) {
	runtime.EventsEmit(a.ctx, "spotify:track", map[string]string{
		"trackName":  trackName,
		"artistName": artistName,
	})
}

// GetSpotifyEnabled returns whether the Spotify integration is enabled.
func (a *App) GetSpotifyEnabled() bool {
	return a.favorites.SpotifyEnabled()
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
		a.spotify.startPolling(a.ctx, auth, token, a.onSpotifyToken, a.onSpotifyTrack)
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
		ReleaseNotes:   info.ReleaseNotes,
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

// launchInstallerAndQuit starts the NSIS installer in a new process group so
// it survives the parent process exiting, then quits the app.
func (a *App) launchInstallerAndQuit(installerPath string) error {
	// No silent flag — the standard NSIS UI shows progress and UAC prompt,
	// giving the user visibility and control during installation.
	cmd := exec.Command(installerPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch installer: %w", err)
	}
	runtime.Quit(a.ctx)
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

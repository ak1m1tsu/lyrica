package main

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"strings"

	"github.com/ak1m1tsu/lyrica/internal/domain"
	"github.com/ak1m1tsu/lyrica/internal/service"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const appVersion = "3.2.0"

// App is the Wails-bound adapter. It owns the Wails runtime context and
// delegates all business logic to the injected services.
type App struct {
	ctx           context.Context
	lyrics        *service.Lyrics
	favorites     *service.Favorites
	discord       *discordPresence
	windowVisible bool
}

// NewApp wires the infrastructure into the services and returns the adapter.
func NewApp(lyrics *service.Lyrics, favorites *service.Favorites) *App {
	return &App{lyrics: lyrics, favorites: favorites, discord: newDiscordPresence()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.windowVisible = true
	if a.favorites.DiscordPresence() {
		a.discord.connect()
		a.discord.setIdle()
	}
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

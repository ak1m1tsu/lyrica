// Package storage provides a SQLite-backed implementation of domain.FavoritesStore.
package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	sq "github.com/Masterminds/squirrel"
	"github.com/ak1m1tsu/lyrica/internal/domain"
	_ "modernc.org/sqlite"
)

const (
	configFileName = "config.json"
	dbFileName     = "lyrica.db"
)

// config holds the persisted application settings.
type config struct {
	FavoritesDir        string `json:"favoritesDir"`
	CloseToTray         bool   `json:"closeToTray"`
	DiscordPresence     bool   `json:"discordPresence"`
	SpotifyEnabled       bool   `json:"spotifyEnabled"`
	SpotifyAccessToken   string `json:"spotifyAccessToken"`
	SpotifyRefreshToken  string `json:"spotifyRefreshToken"`
	GoogleDriveEnabled   bool   `json:"googleDriveEnabled"`
	GoogleAccessToken    string `json:"googleAccessToken"`
	GoogleRefreshToken   string `json:"googleRefreshToken"`
	LastSyncAt           string `json:"lastSyncAt"`
	CurrentTheme         string `json:"currentTheme"`
}

func defaultConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "lyrica")
}

var _ domain.FavoritesStore = (*SQLiteStore)(nil)

// SQLiteStore is a thread-safe, SQLite-backed favorites store.
// The in-memory index provides O(1) Has() lookups without a DB round-trip.
type SQLiteStore struct {
	mu     sync.RWMutex
	db     *sql.DB
	index  map[int]struct{}
	cfg    config
	cfgDir string
}

// NewSQLiteStore opens (or creates) lrclib.db in the configured favorites
// directory and returns a ready store. cfgDir is the directory used for
// config.json; when empty it defaults to the OS user config dir.
func NewSQLiteStore(cfgDir string) (*SQLiteStore, error) {
	if cfgDir == "" {
		cfgDir = defaultConfigDir()
	}
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		return nil, err
	}
	s := &SQLiteStore{
		cfgDir: cfgDir,
		index:  map[int]struct{}{},
	}
	if cfg, err := readConfig(cfgDir); err == nil {
		s.cfg = cfg
	}
	if s.cfg.FavoritesDir == "" {
		s.cfg.FavoritesDir = cfgDir
	}
	if err := os.MkdirAll(s.cfg.FavoritesDir, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(cfgDir, "themes"), 0755); err != nil {
		return nil, err
	}
	db, err := openSQLiteDB(s.dbPath())
	if err != nil {
		return nil, err
	}
	s.db = db
	if err := s.createSchema(); err != nil {
		db.Close()
		return nil, err
	}
	if err := s.loadIndex(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// openSQLiteDB opens the database at path, enables WAL journal mode, and
// constrains the pool to a single connection to avoid locking contention.
func openSQLiteDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func (s *SQLiteStore) dbPath() string {
	return filepath.Join(s.cfg.FavoritesDir, dbFileName)
}

func (s *SQLiteStore) createSchema() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS favorites (
		id            INTEGER PRIMARY KEY,
		track_name    TEXT    NOT NULL DEFAULT '',
		artist_name   TEXT    NOT NULL DEFAULT '',
		album_name    TEXT    NOT NULL DEFAULT '',
		duration      REAL    NOT NULL DEFAULT 0,
		instrumental  INTEGER NOT NULL DEFAULT 0,
		plain_lyrics  TEXT    NOT NULL DEFAULT '',
		synced_lyrics TEXT    NOT NULL DEFAULT ''
	)`)
	return err
}

// loadIndex rebuilds the in-memory ID set from the database.
func (s *SQLiteStore) loadIndex() error {
	rows, err := sq.Select("id").From("favorites").RunWith(s.db).Query()
	if err != nil {
		return err
	}
	defer rows.Close()
	s.index = map[int]struct{}{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return err
		}
		s.index[id] = struct{}{}
	}
	return rows.Err()
}

func readConfig(dir string) (config, error) {
	data, err := os.ReadFile(filepath.Join(dir, configFileName))
	if err != nil {
		return config{}, err
	}
	var cfg config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return config{}, err
	}
	return cfg, nil
}

func (s *SQLiteStore) saveConfig() error {
	data, err := json.MarshalIndent(s.cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.cfgDir, configFileName), data, 0644)
}

// Close closes the underlying database connection. Safe to call multiple times.
func (s *SQLiteStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db == nil {
		return nil
	}
	err := s.db.Close()
	s.db = nil
	return err
}

// GetAll returns a copy of all saved favorites in insertion order.
func (s *SQLiteStore) GetAll() []domain.Track {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.db == nil {
		return []domain.Track{}
	}
	rows, err := sq.Select("id", "track_name", "artist_name", "album_name", "duration", "instrumental", "plain_lyrics", "synced_lyrics").
		From("favorites").
		OrderBy("rowid").
		RunWith(s.db).
		Query()
	if err != nil {
		slog.Error("sqlite: GetAll query", "error", err)
		return []domain.Track{}
	}
	defer rows.Close()
	tracks := []domain.Track{}
	for rows.Next() {
		var t domain.Track
		var instrumental int
		if err := rows.Scan(&t.ID, &t.TrackName, &t.ArtistName, &t.AlbumName, &t.Duration, &instrumental, &t.PlainLyrics, &t.SyncedLyrics); err != nil {
			slog.Error("sqlite: GetAll scan", "error", err)
			continue
		}
		t.Instrumental = instrumental != 0
		tracks = append(tracks, t)
	}
	if err := rows.Err(); err != nil {
		slog.Error("sqlite: GetAll rows", "error", err)
	}
	return tracks
}

// Has reports whether a track with the given ID is a favorite.
func (s *SQLiteStore) Has(id int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.index[id]
	return ok
}

// Add stores a track, deduplicating by ID.
func (s *SQLiteStore) Add(ctx context.Context, track domain.Track) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.index[track.ID]; ok {
		return nil
	}
	instrumental := 0
	if track.Instrumental {
		instrumental = 1
	}
	_, err := sq.Insert("favorites").
		Columns("id", "track_name", "artist_name", "album_name", "duration", "instrumental", "plain_lyrics", "synced_lyrics").
		Values(track.ID, track.TrackName, track.ArtistName, track.AlbumName, track.Duration, instrumental, track.PlainLyrics, track.SyncedLyrics).
		RunWith(s.db).
		ExecContext(ctx)
	if err != nil {
		return err
	}
	s.index[track.ID] = struct{}{}
	return nil
}

// Remove deletes the favorite with the given ID.
func (s *SQLiteStore) Remove(ctx context.Context, id int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.index[id]; !ok {
		return nil
	}
	if _, err := sq.Delete("favorites").
		Where(sq.Eq{"id": id}).
		RunWith(s.db).
		ExecContext(ctx); err != nil {
		return err
	}
	delete(s.index, id)
	return nil
}

// GetDir returns the current favorites storage directory.
func (s *SQLiteStore) GetDir() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.FavoritesDir
}

// SetDir changes the favorites storage directory, copying lrclib.db from the
// old location to the new one.
func (s *SQLiteStore) SetDir(ctx context.Context, newDir string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(newDir, 0755); err != nil {
		return err
	}

	oldDir := s.cfg.FavoritesDir
	s.cfg.FavoritesDir = newDir
	if err := s.saveConfig(); err != nil {
		s.cfg.FavoritesDir = oldDir
		return err
	}

	// Checkpoint WAL into the main file so a plain file copy captures all data.
	s.db.ExecContext(ctx, "PRAGMA wal_checkpoint(FULL)") //nolint:errcheck
	if err := s.db.Close(); err != nil {
		slog.Warn("sqlite: close before SetDir migration", "error", err)
	}
	s.db = nil

	oldPath := filepath.Join(oldDir, dbFileName)
	newPath := s.dbPath()
	if data, err := os.ReadFile(oldPath); err == nil {
		if err := os.WriteFile(newPath, data, 0644); err != nil {
			return err
		}
	}

	db, err := openSQLiteDB(newPath)
	if err != nil {
		return err
	}
	s.db = db
	return nil
}

// GetCloseToTray returns the persisted close-to-tray preference.
func (s *SQLiteStore) GetCloseToTray() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.CloseToTray
}

// SetCloseToTray persists the close-to-tray preference.
func (s *SQLiteStore) SetCloseToTray(ctx context.Context, enabled bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.CloseToTray = enabled
	return s.saveConfig()
}

// GetDiscordPresence returns the persisted Discord Rich Presence preference.
func (s *SQLiteStore) GetDiscordPresence() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.DiscordPresence
}

// SetDiscordPresence persists the Discord Rich Presence preference.
func (s *SQLiteStore) SetDiscordPresence(ctx context.Context, enabled bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.DiscordPresence = enabled
	return s.saveConfig()
}

// GetSpotifyEnabled returns the persisted Spotify integration preference.
func (s *SQLiteStore) GetSpotifyEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.SpotifyEnabled
}

// SetSpotifyEnabled persists the Spotify integration preference.
func (s *SQLiteStore) SetSpotifyEnabled(ctx context.Context, enabled bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.SpotifyEnabled = enabled
	return s.saveConfig()
}

// GetSpotifyAccessToken returns the persisted Spotify access token.
func (s *SQLiteStore) GetSpotifyAccessToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.SpotifyAccessToken
}

// SetSpotifyAccessToken persists the Spotify access token.
func (s *SQLiteStore) SetSpotifyAccessToken(ctx context.Context, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.SpotifyAccessToken = token
	return s.saveConfig()
}

// GetSpotifyRefreshToken returns the persisted Spotify refresh token.
func (s *SQLiteStore) GetSpotifyRefreshToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.SpotifyRefreshToken
}

// SetSpotifyRefreshToken persists the Spotify refresh token.
func (s *SQLiteStore) SetSpotifyRefreshToken(ctx context.Context, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.SpotifyRefreshToken = token
	return s.saveConfig()
}

// GoogleDriveEnabled returns whether Google Drive sync is enabled.
func (s *SQLiteStore) GoogleDriveEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.GoogleDriveEnabled
}

// SetGoogleDriveEnabled persists the Google Drive sync preference.
func (s *SQLiteStore) SetGoogleDriveEnabled(ctx context.Context, enabled bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.GoogleDriveEnabled = enabled
	return s.saveConfig()
}

// GetGoogleAccessToken returns the persisted Google OAuth access token.
func (s *SQLiteStore) GetGoogleAccessToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.GoogleAccessToken
}

// SetGoogleAccessToken persists the Google OAuth access token.
func (s *SQLiteStore) SetGoogleAccessToken(ctx context.Context, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.GoogleAccessToken = token
	return s.saveConfig()
}

// GetGoogleRefreshToken returns the persisted Google OAuth refresh token.
func (s *SQLiteStore) GetGoogleRefreshToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.GoogleRefreshToken
}

// SetGoogleRefreshToken persists the Google OAuth refresh token.
func (s *SQLiteStore) SetGoogleRefreshToken(ctx context.Context, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.GoogleRefreshToken = token
	return s.saveConfig()
}

// GetLastSyncAt returns the RFC3339 timestamp of the last successful sync.
func (s *SQLiteStore) GetLastSyncAt() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.LastSyncAt
}

// SetLastSyncAt persists the last successful sync timestamp.
func (s *SQLiteStore) SetLastSyncAt(ctx context.Context, t string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.LastSyncAt = t
	return s.saveConfig()
}

// GetCurrentTheme returns the ID of the currently active theme.
func (s *SQLiteStore) GetCurrentTheme() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.CurrentTheme
}

// SetCurrentTheme persists the active theme ID.
func (s *SQLiteStore) SetCurrentTheme(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cfg.CurrentTheme = id
	return s.saveConfig()
}

// GetThemesDir returns the directory where custom theme JSON files are stored.
func (s *SQLiteStore) GetThemesDir() string {
	return filepath.Join(s.cfgDir, "themes")
}

// GetCustomThemes reads all *.json files in the themes directory and returns
// the parsed themes, skipping files that fail to parse.
func (s *SQLiteStore) GetCustomThemes() ([]domain.Theme, error) {
	dir := s.GetThemesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []domain.Theme{}, nil
		}
		return nil, err
	}
	themes := []domain.Theme{}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			slog.Warn("themes: skipping unreadable file", "file", e.Name(), "error", err)
			continue
		}
		var t domain.Theme
		if err := json.Unmarshal(data, &t); err != nil {
			slog.Warn("themes: skipping invalid JSON", "file", e.Name(), "error", err)
			continue
		}
		themes = append(themes, t)
	}
	return themes, nil
}

// SaveCustomTheme writes the theme as {id}.json to the themes directory,
// overwriting any existing file with the same ID.
func (s *SQLiteStore) SaveCustomTheme(theme domain.Theme) error {
	data, err := json.MarshalIndent(theme, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.GetThemesDir(), theme.ID+".json"), data, 0644)
}

// DeleteCustomTheme removes {id}.json from the themes directory.
// Returns nil if the file does not exist.
func (s *SQLiteStore) DeleteCustomTheme(id string) error {
	err := os.Remove(filepath.Join(s.GetThemesDir(), id+".json"))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

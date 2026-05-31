// Package storage provides a JSON file-backed implementation of the
// domain.FavoritesStore port.
package storage

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/ak1m1tsu/lrclib/internal/domain"
)

const (
	configFileName    = "config.json"
	favoritesFileName = "favorites.json"
	writeQueueSize    = 16
)

var _ domain.FavoritesStore = (*FileStore)(nil)

// config holds the persisted application settings.
type config struct {
	FavoritesDir string `json:"favoritesDir"`
}

// writeRequest is a single serialized disk write handed to the writer goroutine.
type writeRequest struct {
	path string
	data []byte
}

// FileStore is a thread-safe, JSON file-backed favorites store.
//
// All disk writes for favorites.json are funneled through a single long-lived
// writer goroutine fed by the writes channel. This guarantees writes are
// serialized and executed in submission order (last logical mutation wins),
// and lets Close flush any queued writes before the process exits.
type FileStore struct {
	mu     sync.RWMutex
	tracks []domain.Track
	index  map[int]struct{}
	cfg    config
	cfgDir string
	closed bool

	writes chan writeRequest
	wg     sync.WaitGroup
}

// NewFileStore loads config and favorites from disk and returns a ready store
// with its writer goroutine running. cfgDir is the directory used for
// config.json; when empty it defaults to the OS user config dir. Callers must
// call Close to flush pending writes before exit.
func NewFileStore(cfgDir string) *FileStore {
	if cfgDir == "" {
		cfgDir = defaultConfigDir()
	}
	_ = os.MkdirAll(cfgDir, 0755)
	s := &FileStore{
		cfgDir: cfgDir,
		tracks: []domain.Track{},
		index:  map[int]struct{}{},
		writes: make(chan writeRequest, writeQueueSize),
	}
	s.loadConfig()
	if s.cfg.FavoritesDir == "" {
		s.cfg.FavoritesDir = cfgDir
	}
	s.loadFavorites()

	s.wg.Add(1)
	go s.writeLoop()
	return s
}

// writeLoop is the single writer goroutine. It drains queued writes in order
// and surfaces I/O failures via the logger instead of discarding them.
func (s *FileStore) writeLoop() {
	defer s.wg.Done()
	for req := range s.writes {
		if err := os.MkdirAll(filepath.Dir(req.path), 0755); err != nil {
			slog.Error("favorites: create dir", "path", req.path, "error", err)
			continue
		}
		if err := os.WriteFile(req.path, req.data, 0644); err != nil {
			slog.Error("favorites: write file", "path", req.path, "error", err)
		}
	}
}

// Close stops the writer goroutine and waits for all queued writes to flush.
// It is safe to call multiple times. The composition root must call this
// before the process exits so no favorites mutation is lost.
func (s *FileStore) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	close(s.writes)
	s.mu.Unlock()

	s.wg.Wait()
	return nil
}

// rebuildIndex resets the O(1) membership index from the current tracks
// slice. Callers must hold s.mu (write lock).
func (s *FileStore) rebuildIndex() {
	s.index = make(map[int]struct{}, len(s.tracks))
	for _, t := range s.tracks {
		s.index[t.ID] = struct{}{}
	}
}

func defaultConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "lrclib")
}

func (s *FileStore) favoritesPath() string {
	return filepath.Join(s.cfg.FavoritesDir, favoritesFileName)
}

func (s *FileStore) loadConfig() {
	data, err := os.ReadFile(filepath.Join(s.cfgDir, configFileName))
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &s.cfg)
}

func (s *FileStore) saveConfig() error {
	data, err := json.MarshalIndent(s.cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.cfgDir, configFileName), data, 0644)
}

func (s *FileStore) loadFavorites() {
	data, err := os.ReadFile(s.favoritesPath())
	if err != nil {
		return
	}
	var tracks []domain.Track
	if err := json.Unmarshal(data, &tracks); err == nil && tracks != nil {
		s.tracks = tracks
	}
	s.rebuildIndex()
}

// enqueueSaveLocked marshals the current favorites and submits them to the
// writer goroutine. The caller MUST hold s.mu (write lock); this both
// guarantees a consistent snapshot and serializes against Close so the send
// can never race a closed channel.
func (s *FileStore) enqueueSaveLocked() {
	if s.closed {
		return
	}
	data, err := json.MarshalIndent(s.tracks, "", "  ")
	if err != nil {
		slog.Error("favorites: marshal", "error", err)
		return
	}
	s.writes <- writeRequest{path: s.favoritesPath(), data: data}
}

// GetAll returns a copy of all saved favorites.
func (s *FileStore) GetAll() []domain.Track {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Track, len(s.tracks))
	copy(out, s.tracks)
	return out
}

// Has reports whether a track with the given ID is a favorite.
func (s *FileStore) Has(id int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.index[id]
	return ok
}

// Add stores a track, deduplicating by ID, and queues a persist to disk.
func (s *FileStore) Add(ctx context.Context, track domain.Track) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.index[track.ID]; ok {
		return nil
	}
	s.tracks = append(s.tracks, track)
	s.index[track.ID] = struct{}{}
	s.enqueueSaveLocked()
	return nil
}

// Remove deletes the favorite with the given ID and queues a persist to disk.
func (s *FileStore) Remove(ctx context.Context, id int) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.index[id]; !ok {
		return nil
	}
	for i, t := range s.tracks {
		if t.ID == id {
			s.tracks = append(s.tracks[:i], s.tracks[i+1:]...)
			break
		}
	}
	delete(s.index, id)
	s.enqueueSaveLocked()
	return nil
}

// GetDir returns the current favorites storage directory.
func (s *FileStore) GetDir() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.FavoritesDir
}

// SetDir changes the favorites storage directory, migrating favorites.json
// from the old directory to the new one.
func (s *FileStore) SetDir(ctx context.Context, newDir string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	oldPath := s.favoritesPath()
	s.cfg.FavoritesDir = newDir
	if err := s.saveConfig(); err != nil {
		return err
	}
	if data, err := os.ReadFile(oldPath); err == nil {
		if err := os.MkdirAll(newDir, 0755); err != nil {
			return err
		}
		if err := os.WriteFile(s.favoritesPath(), data, 0644); err != nil {
			return err
		}
	}
	return nil
}

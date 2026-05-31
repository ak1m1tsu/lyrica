package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ak1m1tsu/lrclib/internal/lrclib"
)

const (
	configFileName    = "config.json"
	favoritesFileName = "favorites.json"
)

type appConfig struct {
	FavoritesDir string `json:"favoritesDir"`
}

type favoritesManager struct {
	mu     sync.RWMutex
	tracks []lrclib.Track
	index  map[int]struct{}
	cfg    appConfig
	cfgDir string

	// saveMu serializes the async disk writes so two os.WriteFile calls on
	// favorites.json can never interleave. saveGen is a monotonic counter
	// (guarded by mu) that lets a writer drop its snapshot if a newer
	// mutation has been queued, guaranteeing the last logical mutation wins.
	saveMu  sync.Mutex
	saveGen uint64
}

func newFavoritesManager() *favoritesManager {
	cfgDir := defaultConfigDir()
	_ = os.MkdirAll(cfgDir, 0755)
	fm := &favoritesManager{
		cfgDir: cfgDir,
		tracks: []lrclib.Track{},
		index:  map[int]struct{}{},
	}
	fm.loadConfig()
	if fm.cfg.FavoritesDir == "" {
		fm.cfg.FavoritesDir = cfgDir
	}
	fm.loadFavorites()
	return fm
}

// rebuildIndex resets the O(1) membership index from the current tracks
// slice. Callers must hold fm.mu (write lock).
func (fm *favoritesManager) rebuildIndex() {
	fm.index = make(map[int]struct{}, len(fm.tracks))
	for _, t := range fm.tracks {
		fm.index[t.ID] = struct{}{}
	}
}

func defaultConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "lrclib")
}

func (fm *favoritesManager) favoritesPath() string {
	return filepath.Join(fm.cfg.FavoritesDir, favoritesFileName)
}

func (fm *favoritesManager) loadConfig() {
	data, err := os.ReadFile(filepath.Join(fm.cfgDir, configFileName))
	if err != nil {
		return
	}
	_ = json.Unmarshal(data, &fm.cfg)
}

func (fm *favoritesManager) saveConfig() error {
	data, _ := json.MarshalIndent(fm.cfg, "", "  ")
	return os.WriteFile(filepath.Join(fm.cfgDir, configFileName), data, 0644)
}

func (fm *favoritesManager) loadFavorites() {
	data, err := os.ReadFile(fm.favoritesPath())
	if err != nil {
		return
	}
	var tracks []lrclib.Track
	if err := json.Unmarshal(data, &tracks); err == nil && tracks != nil {
		fm.tracks = tracks
	}
	fm.rebuildIndex()
}

// saveFavorites snapshots the in-memory state under the caller's lock and
// performs the actual disk write asynchronously so that add()/remove()
// return without blocking on disk I/O (and without holding fm.mu during the
// write). The caller MUST hold fm.mu (write lock) when invoking this.
//
// Ordering & safety:
//   - The marshalled bytes and destination path are captured up front so the
//     goroutine never touches fm.tracks, avoiding a data race.
//   - Each call bumps a monotonic generation (fm.saveGen). The writer holds
//     fm.saveMu so writes are fully serialized (never interleaved), and drops
//     its snapshot if a newer generation has since been queued — guaranteeing
//     the last logical mutation is the one that ends up on disk.
//
// Trade-off: because the write happens in a goroutine, the WriteFile error
// can no longer be surfaced to the caller — saveFavorites always returns nil.
func (fm *favoritesManager) saveFavorites() error {
	dir := fm.cfg.FavoritesDir
	path := fm.favoritesPath()
	data, _ := json.MarshalIndent(fm.tracks, "", "  ")

	fm.saveGen++
	gen := fm.saveGen

	go func() {
		fm.saveMu.Lock()
		defer fm.saveMu.Unlock()
		// Drop this write if a newer mutation has already been queued; that
		// newer generation's goroutine will (or already did) persist the
		// authoritative state.
		fm.mu.RLock()
		latest := fm.saveGen
		fm.mu.RUnlock()
		if gen != latest {
			return
		}
		_ = os.MkdirAll(dir, 0755)
		_ = os.WriteFile(path, data, 0644)
	}()
	return nil
}

func (fm *favoritesManager) getAll() []lrclib.Track {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	out := make([]lrclib.Track, len(fm.tracks))
	copy(out, fm.tracks)
	return out
}

func (fm *favoritesManager) has(id int) bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	_, ok := fm.index[id]
	return ok
}

func (fm *favoritesManager) add(track lrclib.Track) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	if _, ok := fm.index[track.ID]; ok {
		return nil
	}
	fm.tracks = append(fm.tracks, track)
	fm.index[track.ID] = struct{}{}
	return fm.saveFavorites()
}

func (fm *favoritesManager) remove(id int) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	if _, ok := fm.index[id]; !ok {
		return nil
	}
	for i, t := range fm.tracks {
		if t.ID == id {
			fm.tracks = append(fm.tracks[:i], fm.tracks[i+1:]...)
			break
		}
	}
	delete(fm.index, id)
	return fm.saveFavorites()
}

func (fm *favoritesManager) getDir() string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.cfg.FavoritesDir
}

func (fm *favoritesManager) setDir(newDir string) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	oldPath := fm.favoritesPath()
	fm.cfg.FavoritesDir = newDir
	if err := fm.saveConfig(); err != nil {
		return err
	}
	if data, err := os.ReadFile(oldPath); err == nil {
		_ = os.MkdirAll(newDir, 0755)
		_ = os.WriteFile(fm.favoritesPath(), data, 0644)
	}
	return nil
}

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

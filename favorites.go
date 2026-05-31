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
	cfg    appConfig
	cfgDir string
}

func newFavoritesManager() *favoritesManager {
	cfgDir := defaultConfigDir()
	_ = os.MkdirAll(cfgDir, 0755)
	fm := &favoritesManager{
		cfgDir: cfgDir,
		tracks: []lrclib.Track{},
	}
	fm.loadConfig()
	if fm.cfg.FavoritesDir == "" {
		fm.cfg.FavoritesDir = cfgDir
	}
	fm.loadFavorites()
	return fm
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
}

func (fm *favoritesManager) saveFavorites() error {
	_ = os.MkdirAll(fm.cfg.FavoritesDir, 0755)
	data, _ := json.MarshalIndent(fm.tracks, "", "  ")
	return os.WriteFile(fm.favoritesPath(), data, 0644)
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
	for _, t := range fm.tracks {
		if t.ID == id {
			return true
		}
	}
	return false
}

func (fm *favoritesManager) add(track lrclib.Track) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	for _, t := range fm.tracks {
		if t.ID == track.ID {
			return nil
		}
	}
	fm.tracks = append(fm.tracks, track)
	return fm.saveFavorites()
}

func (fm *favoritesManager) remove(id int) error {
	fm.mu.Lock()
	defer fm.mu.Unlock()
	for i, t := range fm.tracks {
		if t.ID == id {
			fm.tracks = append(fm.tracks[:i], fm.tracks[i+1:]...)
			return fm.saveFavorites()
		}
	}
	return nil
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

package lrclib

import "errors"

// ErrNotFound is returned when lrclib.net cannot find the requested track.
var ErrNotFound = errors.New("track not found")

// Track holds metadata and lyrics for a music track.
type Track struct {
	ID           int     `json:"id"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}

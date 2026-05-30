package lrclib

import "errors"

// ErrNotFound is returned when lrclib.net cannot find the requested track.
var ErrNotFound = errors.New("track not found")

// Track holds metadata and lyrics for a music track.
type Track struct {
	ID           int
	TrackName    string
	ArtistName   string
	AlbumName    string
	Duration     float64
	Instrumental bool
	PlainLyrics  string
	SyncedLyrics string
}

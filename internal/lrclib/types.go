package lrclib

import "errors"

// ErrNotFound is returned when lrclib.net cannot find the requested track.
var ErrNotFound = errors.New("track not found")

// Track holds metadata and lyrics for a music track from lrclib.net.
type Track struct {
	// ID is the unique lrclib.net identifier.
	ID int `json:"id"`
	// TrackName is the song title.
	TrackName string `json:"trackName"`
	// ArtistName is the performing artist.
	ArtistName string `json:"artistName"`
	// AlbumName is the release album; may be empty.
	AlbumName string `json:"albumName"`
	// Duration is the track length in seconds.
	Duration float64 `json:"duration"`
	// Instrumental is true when the track has no vocals.
	Instrumental bool `json:"instrumental"`
	// PlainLyrics contains unsynced lyrics text; may be empty.
	PlainLyrics string `json:"plainLyrics"`
	// SyncedLyrics contains LRC-format timestamped lyrics; may be empty.
	SyncedLyrics string `json:"syncedLyrics"`
}

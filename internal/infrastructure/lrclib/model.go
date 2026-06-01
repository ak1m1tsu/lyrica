package lrclib

import "github.com/ak1m1tsu/lyrica/internal/domain"

// trackModel mirrors the JSON object returned by the lrclib.net API.
type trackModel struct {
	ID           int     `json:"id"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}

// toTrack maps the API JSON model onto the domain Track type.
func (t trackModel) toTrack() domain.Track {
	return domain.Track{
		ID:           t.ID,
		TrackName:    t.TrackName,
		ArtistName:   t.ArtistName,
		AlbumName:    t.AlbumName,
		Duration:     t.Duration,
		Instrumental: t.Instrumental,
		PlainLyrics:  t.PlainLyrics,
		SyncedLyrics: t.SyncedLyrics,
	}
}

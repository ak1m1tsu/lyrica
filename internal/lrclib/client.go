package lrclib

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://lrclib.net"

// Client retrieves lyrics and track metadata from the lrclib.net API.
type Client struct {
	baseURL string
	http    *http.Client
}

// New returns a Client configured for the public lrclib.net API with a 10s timeout.
func New() *Client {
	return &Client{
		baseURL: defaultBaseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// trackJSON mirrors the JSON object returned by the lrclib.net API.
type trackJSON struct {
	ID           int     `json:"id"`
	TrackName    string  `json:"trackName"`
	ArtistName   string  `json:"artistName"`
	AlbumName    string  `json:"albumName"`
	Duration     float64 `json:"duration"`
	Instrumental bool    `json:"instrumental"`
	PlainLyrics  string  `json:"plainLyrics"`
	SyncedLyrics string  `json:"syncedLyrics"`
}

func (t trackJSON) toTrack() Track {
	return Track{
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

// Search returns tracks matching the given free-text query.
func (c *Client) Search(ctx context.Context, query string) ([]Track, error) {
	u := c.baseURL + "/api/search?q=" + url.QueryEscape(query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("lrclib: build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lrclib: do request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("lrclib: unexpected status %s", resp.Status)
	}

	var raw []trackJSON
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("lrclib: decode response: %w", err)
	}

	tracks := make([]Track, len(raw))
	for i, t := range raw {
		tracks[i] = t.toTrack()
	}
	return tracks, nil
}

// GetByID returns the track with the given lrclib.net identifier.
func (c *Client) GetByID(ctx context.Context, id int) (*Track, error) {
	u := fmt.Sprintf("%s/api/get/%d", c.baseURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("lrclib: build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lrclib: do request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		return nil, fmt.Errorf("lrclib: unexpected status %s", resp.Status)
	}

	var raw trackJSON
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("lrclib: decode response: %w", err)
	}

	track := raw.toTrack()
	return &track, nil
}

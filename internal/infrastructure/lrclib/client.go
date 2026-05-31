// Package lrclib provides an HTTP client for the lrclib.net API that
// implements the domain.LyricsClient port.
package lrclib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ak1m1tsu/lrclib/internal/domain"
)

const (
	defaultBaseURL = "https://lrclib.net"
	userAgent      = "lrclib-desktop/1.0"
	requestTimeout = 10 * time.Second
)

var _ domain.LyricsClient = (*Client)(nil)

// Client retrieves lyrics and track metadata from the lrclib.net API.
type Client struct {
	baseURL string
	http    *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the API base URL, e.g. for tests against an
// httptest.Server.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// New returns a Client configured for the public lrclib.net API with a 10s
// timeout. Options may override defaults such as the base URL.
func New(opts ...Option) *Client {
	c := &Client{
		baseURL: defaultBaseURL,
		http:    &http.Client{Timeout: requestTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Search returns tracks matching the given free-text query.
func (c *Client) Search(ctx context.Context, query string) ([]domain.Track, error) {
	u := c.baseURL + "/api/search?q=" + url.QueryEscape(query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("lrclib: build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lrclib: do request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, domain.ErrNotFound
	case http.StatusTooManyRequests:
		return nil, errors.New("lrclib: rate limited, try again later")
	default:
		return nil, fmt.Errorf("lrclib: unexpected status %s", resp.Status)
	}

	var raw []trackModel
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("lrclib: decode response: %w", err)
	}

	tracks := make([]domain.Track, len(raw))
	for i, t := range raw {
		tracks[i] = t.toTrack()
	}
	return tracks, nil
}

// GetByID returns the track with the given lrclib.net identifier.
func (c *Client) GetByID(ctx context.Context, id int) (*domain.Track, error) {
	u := fmt.Sprintf("%s/api/get/%d", c.baseURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("lrclib: build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("lrclib: do request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, domain.ErrNotFound
	case http.StatusTooManyRequests:
		return nil, errors.New("lrclib: rate limited, try again later")
	default:
		return nil, fmt.Errorf("lrclib: unexpected status %s", resp.Status)
	}

	var raw trackModel
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("lrclib: decode response: %w", err)
	}

	track := raw.toTrack()
	return &track, nil
}

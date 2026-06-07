// Package github provides an HTTP client for the GitHub Releases API.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	defaultBaseURL   = "https://api.github.com/repos/ak1m1tsu/lrclib"
	githubUserAgent  = "lyrica-desktop/1.0"
	requestTimeout   = 30 * time.Second
	githubAPIVersion = "2022-11-28"
)

// ReleaseAsset represents a single file attached to a GitHub release.
type ReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// Release holds the fields from a GitHub Releases API response that Lyrica
// needs for its update flow.
type Release struct {
	TagName string         `json:"tag_name"`
	Body    string         `json:"body"`
	Assets  []ReleaseAsset `json:"assets"`
}

// Client fetches release information from the GitHub Releases API.
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

// New returns a Client configured for the GitHub Releases API with a 30s
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

// LatestRelease fetches the latest published release for the repository.
func (c *Client) LatestRelease(ctx context.Context) (*Release, error) {
	u := c.baseURL + "/releases/latest"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("github: build request: %w", err)
	}
	req.Header.Set("User-Agent", githubUserAgent)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", githubAPIVersion)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: do request: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, fmt.Errorf("github: no releases found")
	case http.StatusForbidden, http.StatusTooManyRequests:
		return nil, fmt.Errorf("github: rate limited, try again later")
	default:
		return nil, fmt.Errorf("github: unexpected status %s", resp.Status)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("github: decode response: %w", err)
	}
	return &release, nil
}

package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fixtureRelease is a realistic GitHub Releases API payload used across
// multiple test cases.
const fixtureRelease = `{
	"tag_name": "v3.4.0",
	"body": "## What's Changed\n- Added auto-update support\n- Fixed crash on startup",
	"assets": [
		{
			"name": "lyrica-3.4.0-amd64-installer.exe",
			"browser_download_url": "https://github.com/ak1m1tsu/lyrica/releases/download/v3.4.0/lyrica-3.4.0-amd64-installer.exe",
			"size": 12345678
		},
		{
			"name": "lyrica-3.4.0-arm64-installer.exe",
			"browser_download_url": "https://github.com/ak1m1tsu/lyrica/releases/download/v3.4.0/lyrica-3.4.0-arm64-installer.exe",
			"size": 11111111
		}
	]
}`

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := New(WithBaseURL(srv.URL))
	return srv, c
}

func TestLatestRelease_Success(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/releases/latest" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("User-Agent") != githubUserAgent {
			t.Errorf("unexpected User-Agent: %s", r.Header.Get("User-Agent"))
		}
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Errorf("unexpected Accept: %s", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fixtureRelease)) //nolint:errcheck
	})

	release, err := c.LatestRelease(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if release == nil {
		t.Fatal("expected non-nil release")
	}
	if release.TagName != "v3.4.0" {
		t.Errorf("TagName: got %q, want %q", release.TagName, "v3.4.0")
	}
	if !strings.Contains(release.Body, "auto-update") {
		t.Errorf("Body missing expected text, got: %q", release.Body)
	}
	if len(release.Assets) != 2 {
		t.Fatalf("expected 2 assets, got %d", len(release.Assets))
	}
	asset := release.Assets[0]
	if asset.Name != "lyrica-3.4.0-amd64-installer.exe" {
		t.Errorf("asset Name: got %q", asset.Name)
	}
	if asset.Size != 12345678 {
		t.Errorf("asset Size: got %d, want 12345678", asset.Size)
	}
	if asset.BrowserDownloadURL == "" {
		t.Error("asset BrowserDownloadURL must not be empty")
	}
}

func TestLatestRelease_404(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := c.LatestRelease(context.Background())
	if err == nil {
		t.Fatal("expected error on 404, got nil")
	}
	if !strings.Contains(err.Error(), "no releases found") {
		t.Errorf("error should mention 'no releases found', got: %v", err)
	}
}

func TestLatestRelease_500(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	_, err := c.LatestRelease(context.Background())
	if err == nil {
		t.Fatal("expected error on 500, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected status") {
		t.Errorf("error should mention 'unexpected status', got: %v", err)
	}
}

func TestLatestRelease_RateLimited_403(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	_, err := c.LatestRelease(context.Background())
	if err == nil {
		t.Fatal("expected error on 403, got nil")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Errorf("error should mention 'rate limited', got: %v", err)
	}
}

func TestLatestRelease_RateLimited_429(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})

	_, err := c.LatestRelease(context.Background())
	if err == nil {
		t.Fatal("expected error on 429, got nil")
	}
	if !strings.Contains(err.Error(), "rate limited") {
		t.Errorf("error should mention 'rate limited', got: %v", err)
	}
}

func TestLatestRelease_MalformedJSON(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"tag_name": "v3.4.0", "assets": [`)) //nolint:errcheck
	})

	_, err := c.LatestRelease(context.Background())
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
	if !strings.Contains(err.Error(), "decode response") {
		t.Errorf("error should mention 'decode response', got: %v", err)
	}
}

func TestLatestRelease_ContextCancellation(t *testing.T) {
	// Use a server that blocks until the client disconnects so the test does
	// not rely on timing.
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately before the request is even sent

	_, err := c.LatestRelease(ctx)
	if err == nil {
		t.Fatal("expected error with cancelled context, got nil")
	}
}

func TestLatestRelease_EmptyAssets(t *testing.T) {
	_, c := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"tag_name":"v3.4.0","body":"notes","assets":[]}`)) //nolint:errcheck
	})

	release, err := c.LatestRelease(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(release.Assets) != 0 {
		t.Errorf("expected 0 assets, got %d", len(release.Assets))
	}
}

func TestWithBaseURL(t *testing.T) {
	c := New(WithBaseURL("http://example.com"))
	if c.baseURL != "http://example.com" {
		t.Errorf("WithBaseURL: got %q, want %q", c.baseURL, "http://example.com")
	}
}

func TestNew_Defaults(t *testing.T) {
	c := New()
	if c.baseURL != defaultBaseURL {
		t.Errorf("default baseURL: got %q, want %q", c.baseURL, defaultBaseURL)
	}
	if c.http == nil {
		t.Error("http client must not be nil")
	}
}

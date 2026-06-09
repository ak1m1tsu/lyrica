package service

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	githubinfra "github.com/ak1m1tsu/lyrica/internal/infrastructure/github"
)

// trustedDownloadHosts is the allowlist of GitHub asset delivery domains.
var trustedDownloadHosts = map[string]bool{
	"objects.githubusercontent.com":  true,
	"releases.githubusercontent.com": true,
	"github.com":                     true,
}

// UpdateInfo holds the result of a successful update check.
type UpdateInfo struct {
	Available     bool
	LatestVersion string
	DownloadURL   string
	InstallerName string
	AssetSize     int64
}

// GithubClient is the port that the Updater depends on for release data.
type GithubClient interface {
	LatestRelease(ctx context.Context) (*githubinfra.Release, error)
}

// Updater compares the running version against the latest GitHub release.
type Updater struct {
	currentVersion string
	github         GithubClient
}

// NewUpdater returns an Updater that will compare releases against
// currentVersion (e.g. "3.3.0").
func NewUpdater(currentVersion string, github GithubClient) *Updater {
	return &Updater{currentVersion: currentVersion, github: github}
}

// CheckForUpdate fetches the latest release from GitHub and returns an
// UpdateInfo describing whether a newer version is available.
//
// It scans release assets for a file matching the pattern
// *-amd64-installer.exe. If no such asset is found, Available is false even
// when the remote tag is newer.
func (u *Updater) CheckForUpdate(ctx context.Context) (*UpdateInfo, error) {
	release, err := u.github.LatestRelease(ctx)
	if err != nil {
		return nil, err
	}

	// Strip leading "v" from tag (e.g. "v3.4.0" → "3.4.0").
	latestVersion := strings.TrimPrefix(release.TagName, "v")

	if !newerThan(latestVersion, u.currentVersion) {
		return &UpdateInfo{Available: false}, nil
	}

	// Find the platform-specific portable binary zip asset.
	suffix := platformAssetSuffix()
	var asset *githubinfra.ReleaseAsset
	for i := range release.Assets {
		a := &release.Assets[i]
		if strings.HasSuffix(a.Name, suffix) {
			asset = a
			break
		}
	}
	if asset == nil {
		return &UpdateInfo{Available: false}, nil
	}

	if err := validateAssetURL(asset.BrowserDownloadURL); err != nil {
		return nil, fmt.Errorf("release asset URL rejected: %w", err)
	}

	return &UpdateInfo{
		Available:     true,
		LatestVersion: latestVersion,
		DownloadURL:   asset.BrowserDownloadURL,
		InstallerName: filepath.Base(asset.Name),
		AssetSize:     asset.Size,
	}, nil
}

func validateAssetURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("malformed asset URL: %w", err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("asset URL must use HTTPS, got %q", u.Scheme)
	}
	if !trustedDownloadHosts[strings.ToLower(u.Hostname())] {
		return fmt.Errorf("asset URL host %q is not in the trusted list", u.Hostname())
	}
	return nil
}

// newerThan returns true when candidate is strictly greater than current,
// comparing MAJOR.MINOR.PATCH segments as integers. Returns false on any
// parse error so that a malformed tag never triggers an update.
func newerThan(candidate, current string) bool {
	cv, err := parseVersion(candidate)
	if err != nil {
		return false
	}
	cu, err := parseVersion(current)
	if err != nil {
		return false
	}
	if cv[0] != cu[0] {
		return cv[0] > cu[0]
	}
	if cv[1] != cu[1] {
		return cv[1] > cu[1]
	}
	return cv[2] > cu[2]
}

// platformAssetSuffix returns the release asset filename suffix for the current OS.
func platformAssetSuffix() string {
	switch runtime.GOOS {
	case "darwin":
		return "-darwin-universal.zip"
	default: // windows
		return "-windows-amd64.zip"
	}
}

func parseVersion(v string) ([3]int, error) {
	parts := strings.SplitN(v, ".", 3)
	if len(parts) != 3 {
		return [3]int{}, strconv.ErrSyntax
	}
	var out [3]int
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return [3]int{}, err
		}
		out[i] = n
	}
	return out, nil
}

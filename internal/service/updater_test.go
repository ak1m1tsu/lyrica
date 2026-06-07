package service

import (
	"context"
	"errors"
	"testing"

	githubinfra "github.com/ak1m1tsu/lyrica/internal/infrastructure/github"
)

// stubGithubClient implements GithubClient for testing. Either returnRelease
// or returnErr is used, depending on which is set.
type stubGithubClient struct {
	release *githubinfra.Release
	err     error
}

func (s *stubGithubClient) LatestRelease(_ context.Context) (*githubinfra.Release, error) {
	return s.release, s.err
}

// windowsAsset is a convenience helper that returns a release asset matching
// the Windows NSIS installer naming pattern.
func windowsAsset(version string) githubinfra.ReleaseAsset {
	name := "lyrica-" + version + "-amd64-installer.exe"
	return githubinfra.ReleaseAsset{
		Name:               name,
		BrowserDownloadURL: "https://objects.githubusercontent.com/repos/ak1m1tsu/lrclib/releases/assets/1/" + name,
		Size:               9999999,
	}
}

// ---- newerThan exhaustive table ----

func TestNewerThan(t *testing.T) {
	cases := []struct {
		name      string
		candidate string
		current   string
		want      bool
	}{
		// Equal versions
		{"equal_patch", "1.0.0", "1.0.0", false},
		{"equal_minor", "2.3.0", "2.3.0", false},
		{"equal_major", "10.0.0", "10.0.0", false},

		// Patch bump
		{"patch_newer", "1.0.1", "1.0.0", true},
		{"patch_older", "1.0.0", "1.0.1", false},

		// Minor bump
		{"minor_newer", "1.1.0", "1.0.9", true},
		{"minor_older", "1.0.0", "1.1.0", false},
		{"minor_same_patch_newer", "2.2.1", "2.2.0", true},

		// Major bump
		{"major_newer", "2.0.0", "1.9.9", true},
		{"major_older", "1.0.0", "2.0.0", false},
		{"major_large", "10.0.0", "9.99.99", true},

		// Malformed — must return false (never trigger update)
		{"malformed_candidate", "not-a-version", "1.0.0", false},
		{"malformed_current", "1.0.0", "not-a-version", false},
		{"both_malformed", "abc", "def", false},
		{"missing_patch", "1.0", "1.0.0", false},
		{"extra_segments", "1.0.0.0", "1.0.0", false},
		{"empty_candidate", "", "1.0.0", false},
		{"empty_current", "1.0.0", "", false},
		{"alpha_segment", "1.0.a", "1.0.0", false},

		// Leading-v prefix — newerThan itself does NOT strip "v"; callers do
		{"v_prefix_candidate", "v3.4.0", "3.3.0", false},
		{"v_prefix_current", "3.4.0", "v3.3.0", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := newerThan(tc.candidate, tc.current)
			if got != tc.want {
				t.Errorf("newerThan(%q, %q) = %v, want %v", tc.candidate, tc.current, got, tc.want)
			}
		})
	}
}

// ---- CheckForUpdate tests ----

func TestCheckForUpdate_UpdateAvailable_WithWindowsAsset(t *testing.T) {
	stub := &stubGithubClient{
		release: &githubinfra.Release{
			TagName: "v3.4.0",
			Body:    "Release notes for 3.4.0",
			Assets:  []githubinfra.ReleaseAsset{windowsAsset("3.4.0")},
		},
	}
	u := NewUpdater("3.3.0", stub)

	info, err := u.CheckForUpdate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.Available {
		t.Fatal("expected Available=true")
	}
	if info.LatestVersion != "3.4.0" {
		t.Errorf("LatestVersion: got %q, want %q", info.LatestVersion, "3.4.0")
	}
	if info.ReleaseNotes != "Release notes for 3.4.0" {
		t.Errorf("ReleaseNotes: got %q", info.ReleaseNotes)
	}
	if info.DownloadURL == "" {
		t.Error("DownloadURL must not be empty")
	}
	if info.InstallerName == "" {
		t.Error("InstallerName must not be empty")
	}
	if info.AssetSize != 9999999 {
		t.Errorf("AssetSize: got %d, want 9999999", info.AssetSize)
	}
}

func TestCheckForUpdate_NoUpdate_SameVersion(t *testing.T) {
	stub := &stubGithubClient{
		release: &githubinfra.Release{
			TagName: "v3.3.0",
			Body:    "same version notes",
			Assets:  []githubinfra.ReleaseAsset{windowsAsset("3.3.0")},
		},
	}
	u := NewUpdater("3.3.0", stub)

	info, err := u.CheckForUpdate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Available {
		t.Error("expected Available=false when remote version equals current")
	}
}

func TestCheckForUpdate_NoUpdate_OlderRemoteVersion(t *testing.T) {
	stub := &stubGithubClient{
		release: &githubinfra.Release{
			TagName: "v3.2.1",
			Body:    "older release",
			Assets:  []githubinfra.ReleaseAsset{windowsAsset("3.2.1")},
		},
	}
	u := NewUpdater("3.3.0", stub)

	info, err := u.CheckForUpdate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Available {
		t.Error("expected Available=false when remote version is older than current")
	}
}

func TestCheckForUpdate_NoUpdate_NoWindowsAsset(t *testing.T) {
	stub := &stubGithubClient{
		release: &githubinfra.Release{
			TagName: "v3.4.0",
			Body:    "notes",
			Assets: []githubinfra.ReleaseAsset{
				{
					Name:               "lyrica-3.4.0-checksums.txt",
					BrowserDownloadURL: "https://example.com/checksums.txt",
					Size:               512,
				},
				{
					Name:               "lyrica-3.4.0-source.tar.gz",
					BrowserDownloadURL: "https://example.com/source.tar.gz",
					Size:               5000000,
				},
			},
		},
	}
	u := NewUpdater("3.3.0", stub)

	info, err := u.CheckForUpdate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Available {
		t.Error("expected Available=false when no Windows installer asset is present")
	}
}

func TestCheckForUpdate_NoUpdate_EmptyAssets(t *testing.T) {
	stub := &stubGithubClient{
		release: &githubinfra.Release{
			TagName: "v3.4.0",
			Body:    "notes",
			Assets:  []githubinfra.ReleaseAsset{},
		},
	}
	u := NewUpdater("3.3.0", stub)

	info, err := u.CheckForUpdate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Available {
		t.Error("expected Available=false for empty assets list")
	}
}

func TestCheckForUpdate_ClientError(t *testing.T) {
	expectedErr := errors.New("github: rate limited, try again later")
	stub := &stubGithubClient{err: expectedErr}
	u := NewUpdater("3.3.0", stub)

	info, err := u.CheckForUpdate(context.Background())
	if err == nil {
		t.Fatal("expected error from client, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected wrapped client error, got: %v", err)
	}
	if info != nil {
		t.Errorf("expected nil UpdateInfo on error, got %+v", info)
	}
}

func TestCheckForUpdate_StripsLeadingV_FromTag(t *testing.T) {
	stub := &stubGithubClient{
		release: &githubinfra.Release{
			TagName: "v4.0.0",
			Body:    "",
			Assets:  []githubinfra.ReleaseAsset{windowsAsset("4.0.0")},
		},
	}
	u := NewUpdater("3.3.0", stub)

	info, err := u.CheckForUpdate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.Available {
		t.Fatal("expected Available=true")
	}
	// LatestVersion must have the "v" stripped.
	if info.LatestVersion != "4.0.0" {
		t.Errorf("LatestVersion must not contain leading 'v', got %q", info.LatestVersion)
	}
}

func TestCheckForUpdate_SelectsFirstWindowsAsset(t *testing.T) {
	// Release has multiple assets; the first matching the installer suffix wins.
	stub := &stubGithubClient{
		release: &githubinfra.Release{
			TagName: "v3.4.0",
			Body:    "",
			Assets: []githubinfra.ReleaseAsset{
				{
					Name:               "lyrica-3.4.0-amd64-installer.exe",
					BrowserDownloadURL: "https://objects.githubusercontent.com/repos/ak1m1tsu/lrclib/releases/assets/1/lyrica-3.4.0-amd64-installer.exe",
					Size:               1111,
				},
				{
					Name:               "lyrica-3.4.0-arm64-installer.exe",
					BrowserDownloadURL: "https://objects.githubusercontent.com/repos/ak1m1tsu/lrclib/releases/assets/2/lyrica-3.4.0-arm64-installer.exe",
					Size:               2222,
				},
			},
		},
	}
	u := NewUpdater("3.3.0", stub)

	info, err := u.CheckForUpdate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.Available {
		t.Fatal("expected Available=true")
	}
	if info.DownloadURL != "https://objects.githubusercontent.com/repos/ak1m1tsu/lrclib/releases/assets/1/lyrica-3.4.0-amd64-installer.exe" {
		t.Errorf("expected first matching asset, got DownloadURL=%q", info.DownloadURL)
	}
	if info.AssetSize != 1111 {
		t.Errorf("expected AssetSize=1111, got %d", info.AssetSize)
	}
}

func TestCheckForUpdate_InstallerName_BaseNameOnly(t *testing.T) {
	stub := &stubGithubClient{
		release: &githubinfra.Release{
			TagName: "v3.4.0",
			Body:    "",
			Assets: []githubinfra.ReleaseAsset{
				{
					Name:               "lyrica-3.4.0-amd64-installer.exe",
					BrowserDownloadURL: "https://objects.githubusercontent.com/repos/ak1m1tsu/lrclib/releases/assets/1/lyrica-3.4.0-amd64-installer.exe",
					Size:               5000,
				},
			},
		},
	}
	u := NewUpdater("3.3.0", stub)

	info, err := u.CheckForUpdate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// InstallerName should be just the file name, not a full URL path.
	if info.InstallerName != "lyrica-3.4.0-amd64-installer.exe" {
		t.Errorf("InstallerName: got %q, want %q", info.InstallerName, "lyrica-3.4.0-amd64-installer.exe")
	}
}

func TestCheckForUpdate_MajorVersionBump(t *testing.T) {
	stub := &stubGithubClient{
		release: &githubinfra.Release{
			TagName: "v4.0.0",
			Body:    "major release",
			Assets:  []githubinfra.ReleaseAsset{windowsAsset("4.0.0")},
		},
	}
	u := NewUpdater("3.99.99", stub)

	info, err := u.CheckForUpdate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.Available {
		t.Error("expected Available=true for major version bump")
	}
}

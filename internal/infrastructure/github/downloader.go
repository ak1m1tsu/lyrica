package github

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	downloadChunkSize = 32 * 1024         // 32 KiB
	maxInstallerBytes = 200 * 1024 * 1024 // 200 MiB hard ceiling
)

// ProgressFunc is called after each chunk is written to disk.
// received is the number of bytes written so far; total is the Content-Length
// reported by the server (may be 0 if unknown).
type ProgressFunc func(received, total int64)

// DownloadBinary downloads the binary zip at the given URL to a temporary
// file and returns its absolute path. Progress is reported via the progress
// callback after each 32 KiB chunk.
//
// If the context is cancelled or any error occurs the partial file is removed
// before returning.
func DownloadBinary(ctx context.Context, url string, progress ProgressFunc) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("github: build download request: %w", err)
	}
	req.Header.Set("User-Agent", githubUserAgent)

	// No timeout — files can be large; caller controls cancellation via context.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("github: download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github: unexpected download status %s", resp.Status)
	}

	if resp.ContentLength > maxInstallerBytes {
		return "", fmt.Errorf("github: binary too large (%d bytes)", resp.ContentLength)
	}

	tmp, err := os.CreateTemp("", "lyrica-update-*.zip")
	if err != nil {
		return "", fmt.Errorf("github: create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	// cleanup helper — called on any error path
	cleanup := func() {
		tmp.Close()
		os.Remove(tmpPath)
	}

	total := max(resp.ContentLength, 0) // ContentLength is -1 when unknown; treat as 0 for UI

	buf := make([]byte, downloadChunkSize)
	var received int64
	limited := io.LimitReader(resp.Body, maxInstallerBytes+1)

	for {
		n, readErr := limited.Read(buf)
		if n > 0 {
			if _, writeErr := tmp.Write(buf[:n]); writeErr != nil {
				cleanup()
				return "", fmt.Errorf("github: write temp file: %w", writeErr)
			}
			received += int64(n)
			if received > maxInstallerBytes {
				cleanup()
				return "", fmt.Errorf("github: installer exceeds %d MiB limit", maxInstallerBytes/1024/1024)
			}
			if progress != nil {
				progress(received, total)
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			cleanup()
			return "", fmt.Errorf("github: read download stream: %w", readErr)
		}
		// Check for context cancellation between chunks.
		select {
		case <-ctx.Done():
			cleanup()
			return "", fmt.Errorf("github: download cancelled: %w", ctx.Err())
		default:
		}
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("github: close temp file: %w", err)
	}
	return tmpPath, nil
}

// ExtractBinaryFromZip opens a zip archive downloaded by DownloadBinary and
// extracts the platform executable to a separate temp file, returning its path.
// On Windows it looks for the first .exe entry; on macOS it looks for the
// Contents/MacOS/ binary inside the .app bundle.
func ExtractBinaryFromZip(zipPath string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("github: open zip: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if !isBinaryEntry(f.Name) {
			continue
		}
		src, err := f.Open()
		if err != nil {
			return "", fmt.Errorf("github: open zip entry %q: %w", f.Name, err)
		}

		ext := filepath.Ext(f.Name)
		tmp, err := os.CreateTemp("", "lyrica-bin-*"+ext)
		if err != nil {
			src.Close()
			return "", fmt.Errorf("github: create temp binary: %w", err)
		}
		tmpPath := tmp.Name()

		limited := io.LimitReader(src, maxInstallerBytes+1)
		if _, err := io.Copy(tmp, limited); err != nil {
			src.Close()
			tmp.Close()
			os.Remove(tmpPath)
			return "", fmt.Errorf("github: extract binary: %w", err)
		}
		src.Close()
		if err := tmp.Close(); err != nil {
			os.Remove(tmpPath)
			return "", fmt.Errorf("github: close extracted binary: %w", err)
		}
		// Ensure the extracted file is executable (no-op on Windows).
		_ = os.Chmod(tmpPath, 0755)
		return tmpPath, nil
	}
	return "", fmt.Errorf("github: no binary found in zip")
}

// isBinaryEntry returns true for zip entries that are the main executable.
func isBinaryEntry(name string) bool {
	switch runtime.GOOS {
	case "darwin":
		// Match Contents/MacOS/<name> — the actual binary inside the .app bundle.
		return strings.Contains(name, "Contents/MacOS/") && !strings.HasSuffix(name, "/")
	default: // windows
		return strings.HasSuffix(name, ".exe") && !strings.Contains(name, "/")
	}
}

package github

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	downloadChunkSize   = 32 * 1024        // 32 KiB
	maxInstallerBytes   = 200 * 1024 * 1024 // 200 MiB hard ceiling
)

// ProgressFunc is called after each chunk is written to disk.
// received is the number of bytes written so far; total is the Content-Length
// reported by the server (may be 0 if unknown).
type ProgressFunc func(received, total int64)

// DownloadInstaller downloads the installer at the given URL to a temporary
// file and returns its absolute path. Progress is reported via the progress
// callback after each 32 KiB chunk.
//
// If the context is cancelled or any error occurs the partial file is removed
// before returning.
func DownloadInstaller(ctx context.Context, url string, progress ProgressFunc) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("github: build download request: %w", err)
	}
	req.Header.Set("User-Agent", githubUserAgent)

	// No timeout — installer files can be large; caller controls cancellation
	// via the context.
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
		return "", fmt.Errorf("github: installer too large (%d bytes)", resp.ContentLength)
	}

	tmp, err := os.CreateTemp("", "lyrica-update-*.exe")
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

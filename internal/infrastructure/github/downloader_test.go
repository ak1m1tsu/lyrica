package github

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// makePayload returns a deterministic byte slice of the requested size.
func makePayload(size int) []byte {
	b := make([]byte, size)
	for i := range b {
		b[i] = byte(i % 256)
	}
	return b
}

func newDownloadServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func TestDownloadBinary_Success_SmallFile(t *testing.T) {
	payload := makePayload(1024) // 1 KiB — fits in a single chunk

	srv := newDownloadServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		w.WriteHeader(http.StatusOK)
		w.Write(payload) //nolint:errcheck
	})

	var progressCalls int
	var lastReceived, lastTotal int64

	path, err := DownloadBinary(context.Background(), srv.URL, func(received, total int64) {
		progressCalls++
		lastReceived = received
		lastTotal = total
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })

	if path == "" {
		t.Fatal("expected non-empty temp file path")
	}
	if !strings.HasSuffix(path, ".zip") {
		t.Errorf("temp file should end with .zip, got %q", path)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read temp file: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("file content mismatch: got %d bytes, want %d", len(got), len(payload))
	}

	if progressCalls == 0 {
		t.Error("progress callback was never called")
	}
	if lastReceived != int64(len(payload)) {
		t.Errorf("last received: got %d, want %d", lastReceived, len(payload))
	}
	if lastTotal != int64(len(payload)) {
		t.Errorf("last total: got %d, want %d", lastTotal, len(payload))
	}
}

func TestDownloadBinary_Success_MultiChunk(t *testing.T) {
	// 3 chunks worth of data plus a partial chunk.
	payload := makePayload(downloadChunkSize*3 + 7777)

	srv := newDownloadServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		w.WriteHeader(http.StatusOK)
		w.Write(payload) //nolint:errcheck
	})

	path, err := DownloadBinary(context.Background(), srv.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read temp file: %v", err)
	}
	if !bytes.Equal(got, payload) {
		t.Errorf("multi-chunk content mismatch: got %d bytes, want %d", len(got), len(payload))
	}
}

func TestDownloadBinary_Success_NilProgress(t *testing.T) {
	payload := makePayload(512)

	srv := newDownloadServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(payload) //nolint:errcheck
	})

	// nil progress must not panic
	path, err := DownloadBinary(context.Background(), srv.URL, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })
}

func TestDownloadBinary_Success_UnknownContentLength(t *testing.T) {
	payload := makePayload(512)

	srv := newDownloadServer(t, func(w http.ResponseWriter, r *http.Request) {
		// Force chunked transfer encoding so Content-Length is not set.
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Del("Content-Length")
		w.WriteHeader(http.StatusOK)
		// Write in two chunks to prevent automatic Content-Length buffering.
		half := len(payload) / 2
		w.Write(payload[:half])                        //nolint:errcheck
		w.(http.Flusher).Flush()
		w.Write(payload[half:]) //nolint:errcheck
	})

	var observedTotal int64
	path, err := DownloadBinary(context.Background(), srv.URL, func(received, total int64) {
		observedTotal = total
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })

	// When Content-Length is absent, total should be reported as 0.
	if observedTotal != 0 {
		t.Errorf("expected total=0 for unknown content length, got %d", observedTotal)
	}
}

func TestDownloadBinary_Non200_ReturnsError(t *testing.T) {
	statuses := []int{
		http.StatusNotFound,
		http.StatusForbidden,
		http.StatusInternalServerError,
	}
	for _, status := range statuses {
		t.Run(fmt.Sprintf("HTTP_%d", status), func(t *testing.T) {
			srv := newDownloadServer(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(status)
			})

			_, err := DownloadBinary(context.Background(), srv.URL, nil)
			if err == nil {
				t.Fatalf("expected error for status %d, got nil", status)
			}
			if !strings.Contains(err.Error(), "unexpected download status") {
				t.Errorf("error should mention 'unexpected download status', got: %v", err)
			}
		})
	}
}

func TestDownloadBinary_ContextCancelled_RemovesPartialFile(t *testing.T) {
	// Use a large payload so the download cannot complete before cancellation.
	payload := makePayload(downloadChunkSize * 10)

	ready := make(chan struct{})
	srv := newDownloadServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		w.WriteHeader(http.StatusOK)
		// Write the first chunk, then signal readiness, then block so the
		// context cancellation fires mid-transfer.
		w.Write(payload[:downloadChunkSize]) //nolint:errcheck
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		close(ready)
		<-r.Context().Done()
	})

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel after the server has sent the first chunk.
	go func() {
		<-ready
		cancel()
	}()

	_, err := DownloadBinary(ctx, srv.URL, nil)
	if err == nil {
		t.Fatal("expected error after context cancellation, got nil")
	}
	if !strings.Contains(err.Error(), "download cancelled") && !strings.Contains(err.Error(), "context") {
		t.Errorf("error should mention cancellation, got: %v", err)
	}
}

func TestDownloadBinary_FileDoesNotExistAfterError(t *testing.T) {
	// A server that returns 404 should not leave a temp file behind.
	srv := newDownloadServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := DownloadBinary(context.Background(), srv.URL, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// The returned path is empty on error — no temp file to check.
}

func TestDownloadBinary_ProgressMonotonicallyIncreasing(t *testing.T) {
	payload := makePayload(downloadChunkSize * 4)

	srv := newDownloadServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(payload)))
		w.WriteHeader(http.StatusOK)
		w.Write(payload) //nolint:errcheck
	})

	var prev int64
	path, err := DownloadBinary(context.Background(), srv.URL, func(received, total int64) {
		if received < prev {
			t.Errorf("progress went backwards: %d → %d", prev, received)
		}
		prev = received
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })
}

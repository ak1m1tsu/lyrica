package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ak1m1tsu/lrclib/internal/lrclib"
	"github.com/gorilla/mux"
)

type mockLyricsClient struct {
	track *lrclib.Track
	err   error
}

func (m *mockLyricsClient) GetByID(_ context.Context, _ int) (*lrclib.Track, error) {
	return m.track, m.err
}

type mockLyricsRenderer struct {
	renderLyricsContentCalled bool
	renderLyricsPageCalled    bool
	renderErrorCalled         bool
	lastPlain                 bool
	errToReturn               error
}

func (m *mockLyricsRenderer) RenderLyricsContent(_ context.Context, w io.Writer, _ lrclib.Track, plain bool) error {
	m.renderLyricsContentCalled = true
	m.lastPlain = plain
	_, _ = io.WriteString(w, "lyrics content")
	return m.errToReturn
}

func (m *mockLyricsRenderer) RenderLyricsPage(_ context.Context, w io.Writer, _ lrclib.Track, plain bool) error {
	m.renderLyricsPageCalled = true
	m.lastPlain = plain
	_, _ = io.WriteString(w, "lyrics page")
	return m.errToReturn
}

func (m *mockLyricsRenderer) RenderError(_ context.Context, w io.Writer, _ string) error {
	m.renderErrorCalled = true
	_, _ = io.WriteString(w, "error block")
	return m.errToReturn
}

func newLyricsRequest(method, target string, vars map[string]string) *http.Request {
	r := httptest.NewRequest(method, target, nil)
	return mux.SetURLVars(r, vars)
}

func TestLyricsHandler_InvalidID_NonHTMX_Returns400(t *testing.T) {
	client := &mockLyricsClient{}
	tmpl := &mockLyricsRenderer{}

	h := LyricsHandler(client, tmpl)
	r := newLyricsRequest(http.MethodGet, "/lyrics/abc", map[string]string{"id": "abc"})
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	if !tmpl.renderErrorCalled {
		t.Error("expected RenderError to be called for invalid ID")
	}
}

func TestLyricsHandler_InvalidID_HTMX_Returns200(t *testing.T) {
	client := &mockLyricsClient{}
	tmpl := &mockLyricsRenderer{}

	h := LyricsHandler(client, tmpl)
	r := newLyricsRequest(http.MethodGet, "/lyrics/abc", map[string]string{"id": "abc"})
	r.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("HTMX invalid ID: status = %d, want %d", w.Code, http.StatusOK)
	}
	if !tmpl.renderErrorCalled {
		t.Error("expected RenderError to be called for invalid ID on HTMX request")
	}
}

func TestLyricsHandler_ValidID_NotFound_NonHTMX_Returns404(t *testing.T) {
	client := &mockLyricsClient{err: lrclib.ErrNotFound}
	tmpl := &mockLyricsRenderer{}

	h := LyricsHandler(client, tmpl)
	r := newLyricsRequest(http.MethodGet, "/lyrics/999", map[string]string{"id": "999"})
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
	if !tmpl.renderErrorCalled {
		t.Error("expected RenderError for ErrNotFound")
	}
}

func TestLyricsHandler_ValidID_NotFound_HTMX_Returns200(t *testing.T) {
	client := &mockLyricsClient{err: lrclib.ErrNotFound}
	tmpl := &mockLyricsRenderer{}

	h := LyricsHandler(client, tmpl)
	r := newLyricsRequest(http.MethodGet, "/lyrics/999", map[string]string{"id": "999"})
	r.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("HTMX not found: status = %d, want %d", w.Code, http.StatusOK)
	}
	if !tmpl.renderErrorCalled {
		t.Error("expected RenderError for ErrNotFound on HTMX request")
	}
}

func TestLyricsHandler_ValidID_Returns200(t *testing.T) {
	track := &lrclib.Track{ID: 42, TrackName: "Song", SyncedLyrics: "[00:01.00] Line"}
	client := &mockLyricsClient{track: track}
	tmpl := &mockLyricsRenderer{}

	h := LyricsHandler(client, tmpl)
	r := newLyricsRequest(http.MethodGet, "/lyrics/42", map[string]string{"id": "42"})
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !tmpl.renderLyricsPageCalled {
		t.Error("expected RenderLyricsPage for valid non-HTMX request")
	}
}

func TestLyricsHandler_PlainParam_PassedToRenderer(t *testing.T) {
	track := &lrclib.Track{ID: 42, TrackName: "Song", PlainLyrics: "plain text"}
	client := &mockLyricsClient{track: track}
	tmpl := &mockLyricsRenderer{}

	h := LyricsHandler(client, tmpl)
	r := newLyricsRequest(http.MethodGet, "/lyrics/42?plain=true", map[string]string{"id": "42"})
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !tmpl.lastPlain {
		t.Error("expected plain=true to be passed to renderer")
	}
}

func TestLyricsHandler_ValidID_HTMX_ReturnsPartial(t *testing.T) {
	track := &lrclib.Track{ID: 42, TrackName: "Song"}
	client := &mockLyricsClient{track: track}
	tmpl := &mockLyricsRenderer{}

	h := LyricsHandler(client, tmpl)
	r := newLyricsRequest(http.MethodGet, "/lyrics/42", map[string]string{"id": "42"})
	r.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !tmpl.renderLyricsContentCalled {
		t.Error("expected RenderLyricsContent for HTMX request")
	}
	if tmpl.renderLyricsPageCalled {
		t.Error("RenderLyricsPage should not be called for HTMX request")
	}
}

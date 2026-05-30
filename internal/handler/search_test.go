package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ak1m1tsu/lrclib/internal/lrclib"
)

type mockSearchClient struct {
	tracks []lrclib.Track
	err    error
	called bool
}

func (m *mockSearchClient) Search(_ context.Context, _ string) ([]lrclib.Track, error) {
	m.called = true
	return m.tracks, m.err
}

type mockSearchRenderer struct {
	renderSearchResultsCalled bool
	renderSearchPageCalled    bool
	renderErrorCalled         bool
	errToReturn               error
}

func (m *mockSearchRenderer) RenderSearchResults(_ context.Context, w io.Writer, _ []lrclib.Track) error {
	m.renderSearchResultsCalled = true
	_, _ = io.WriteString(w, "partial results")
	return m.errToReturn
}

func (m *mockSearchRenderer) RenderSearchPage(_ context.Context, w io.Writer, _ string, _ []lrclib.Track) error {
	m.renderSearchPageCalled = true
	_, _ = io.WriteString(w, "full page")
	return m.errToReturn
}

func (m *mockSearchRenderer) RenderError(_ context.Context, w io.Writer, _ string) error {
	m.renderErrorCalled = true
	_, _ = io.WriteString(w, "error block")
	return m.errToReturn
}

func TestSearchHandler_EmptyQuery(t *testing.T) {
	client := &mockSearchClient{}
	tmpl := &mockSearchRenderer{}

	h := SearchHandler(client, tmpl)
	r := httptest.NewRequest(http.MethodGet, "/search", nil)
	w := httptest.NewRecorder()
	h(w, r)

	if client.called {
		t.Error("client.Search should not be called for empty query")
	}
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestSearchHandler_HTMXRequest_ReturnsPartial(t *testing.T) {
	client := &mockSearchClient{tracks: []lrclib.Track{{ID: 1, TrackName: "Song"}}}
	tmpl := &mockSearchRenderer{}

	h := SearchHandler(client, tmpl)
	r := httptest.NewRequest(http.MethodGet, "/search?q=song", nil)
	r.Header.Set("HX-Request", "true")
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !tmpl.renderSearchResultsCalled {
		t.Error("expected RenderSearchResults to be called for HTMX request")
	}
	if tmpl.renderSearchPageCalled {
		t.Error("RenderSearchPage should not be called for HTMX request")
	}
	if strings.Contains(w.Body.String(), "full page") {
		t.Error("HTMX partial response should not contain full page content")
	}
}

func TestSearchHandler_NonHTMXRequest_ReturnsFullPage(t *testing.T) {
	client := &mockSearchClient{tracks: []lrclib.Track{{ID: 1, TrackName: "Song"}}}
	tmpl := &mockSearchRenderer{}

	h := SearchHandler(client, tmpl)
	r := httptest.NewRequest(http.MethodGet, "/search?q=song", nil)
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !tmpl.renderSearchPageCalled {
		t.Error("expected RenderSearchPage to be called for non-HTMX request")
	}
}

func TestSearchHandler_ClientErrNotFound_RendersEmptyState(t *testing.T) {
	client := &mockSearchClient{err: lrclib.ErrNotFound}
	tmpl := &mockSearchRenderer{}

	h := SearchHandler(client, tmpl)
	r := httptest.NewRequest(http.MethodGet, "/search?q=unknown", nil)
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if tmpl.renderErrorCalled {
		t.Error("RenderError should not be called when ErrNotFound")
	}
	if !tmpl.renderSearchPageCalled {
		t.Error("expected RenderSearchPage with empty tracks for ErrNotFound")
	}
}

func TestSearchHandler_ClientOtherError_RendersErrorBlock(t *testing.T) {
	client := &mockSearchClient{err: errors.New("network failure")}
	tmpl := &mockSearchRenderer{}

	h := SearchHandler(client, tmpl)
	r := httptest.NewRequest(http.MethodGet, "/search?q=song", nil)
	w := httptest.NewRecorder()
	h(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !tmpl.renderErrorCalled {
		t.Error("expected RenderError to be called on client error")
	}
}

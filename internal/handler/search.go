package handler

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/ak1m1tsu/lrclib/internal/lrclib"
)

// searchClient searches lrclib.net for tracks matching a query.
type searchClient interface {
	Search(ctx context.Context, query string) ([]lrclib.Track, error)
}

// searchRenderer renders search results either as a partial or as a full page.
type searchRenderer interface {
	RenderSearchResults(ctx context.Context, w io.Writer, tracks []lrclib.Track) error
	RenderSearchPage(ctx context.Context, w io.Writer, query string, tracks []lrclib.Track) error
	RenderError(ctx context.Context, w io.Writer, msg string) error
}

// SearchHandler returns an http.HandlerFunc that searches for tracks and renders
// the results. HTMX requests receive only the results partial; other requests
// receive the full search page.
func SearchHandler(client searchClient, tmpl searchRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		query := r.URL.Query().Get("q")
		if len(query) > 500 {
			query = query[:500]
		}
		partial := r.Header.Get("HX-Request") == "true"

		render := func(tracks []lrclib.Track) {
			var err error
			if partial {
				err = tmpl.RenderSearchResults(r.Context(), w, tracks)
			} else {
				err = tmpl.RenderSearchPage(r.Context(), w, query, tracks)
			}
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}

		if query == "" {
			render(nil)
			return
		}

		tracks, err := client.Search(r.Context(), query)
		if err != nil {
			if errors.Is(err, lrclib.ErrNotFound) {
				render(nil)
				return
			}
			if rerr := tmpl.RenderError(r.Context(), w, "Search failed. Please try again."); rerr != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		render(tracks)
	}
}

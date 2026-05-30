package handler

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/ak1m1tsu/lrclib/internal/lrclib"
	"github.com/gorilla/mux"
)

// lyricsClient retrieves a single track by its lrclib.net identifier.
type lyricsClient interface {
	GetByID(ctx context.Context, id int) (*lrclib.Track, error)
}

// lyricsRenderer renders lyrics either as a partial or as a full page.
type lyricsRenderer interface {
	RenderLyricsContent(ctx context.Context, w io.Writer, track lrclib.Track, plain bool) error
	RenderLyricsPage(ctx context.Context, w io.Writer, track lrclib.Track, plain bool) error
	RenderError(ctx context.Context, w io.Writer, msg string) error
}

// LyricsHandler returns an http.HandlerFunc that fetches a track by ID and
// renders its lyrics. HTMX requests receive only the lyrics content partial;
// other requests receive the full page.
//
// Error responses always use HTTP 200 for HTMX requests so that HTMX swaps
// the error block into the target. Non-HTMX requests use proper status codes.
func LyricsHandler(client lyricsClient, tmpl lyricsRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		partial := r.Header.Get("HX-Request") == "true"

		writeError := func(status int, msg string) {
			if !partial {
				w.WriteHeader(status)
			}
			_ = tmpl.RenderError(r.Context(), w, msg)
		}

		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			writeError(http.StatusBadRequest, "Invalid track ID.")
			return
		}

		plain := r.URL.Query().Get("plain") == "true"

		track, err := client.GetByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, lrclib.ErrNotFound) {
				writeError(http.StatusNotFound, "Track not found.")
				return
			}
			writeError(http.StatusInternalServerError, "Failed to load lyrics. Please try again.")
			return
		}

		if partial {
			err = tmpl.RenderLyricsContent(r.Context(), w, *track, plain)
		} else {
			err = tmpl.RenderLyricsPage(r.Context(), w, *track, plain)
		}
		if err != nil {
			// Headers already sent; log only.
			_ = err
		}
	}
}

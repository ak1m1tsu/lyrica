package handler

import (
	"context"
	"io"
	"net/http"
)

// homeRenderer renders the home page.
type homeRenderer interface {
	RenderHome(ctx context.Context, w io.Writer) error
}

// HomeHandler returns an http.HandlerFunc that renders the home page.
func HomeHandler(tmpl homeRenderer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.RenderHome(r.Context(), w); err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

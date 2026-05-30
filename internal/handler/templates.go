package handler

import (
	"context"
	"io"

	"github.com/ak1m1tsu/lrclib/internal/lrclib"
	"github.com/ak1m1tsu/lrclib/internal/template"
)

// Templates adapts the generated Templ components to the renderer interfaces
// consumed by the handlers.
type Templates struct{}

// RenderHome renders the home page.
func (Templates) RenderHome(ctx context.Context, w io.Writer) error {
	return template.Home().Render(ctx, w)
}

// RenderSearchResults renders the search results partial.
func (Templates) RenderSearchResults(ctx context.Context, w io.Writer, tracks []lrclib.Track) error {
	return template.SearchResults(tracks).Render(ctx, w)
}

// RenderSearchPage renders the full search page.
func (Templates) RenderSearchPage(ctx context.Context, w io.Writer, query string, tracks []lrclib.Track) error {
	return template.SearchPage(query, tracks).Render(ctx, w)
}

// RenderLyricsContent renders the lyrics content partial.
func (Templates) RenderLyricsContent(ctx context.Context, w io.Writer, track lrclib.Track, plain bool) error {
	return template.LyricsContent(track, plain).Render(ctx, w)
}

// RenderLyricsPage renders the full lyrics page.
func (Templates) RenderLyricsPage(ctx context.Context, w io.Writer, track lrclib.Track, plain bool) error {
	return template.LyricsPage(track, plain).Render(ctx, w)
}

// RenderError renders a reusable error block.
func (Templates) RenderError(ctx context.Context, w io.Writer, msg string) error {
	return template.ErrorBlock(msg).Render(ctx, w)
}

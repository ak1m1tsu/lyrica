// Package service contains the application use cases. It depends only on the
// domain layer and orchestrates domain ports.
package service

import (
	"context"
	"errors"
	"strings"

	"github.com/ak1m1tsu/lyrica/internal/domain"
)

const maxQueryLength = 500

// ErrInvalidID is returned by GetByID when the supplied identifier is not positive.
var ErrInvalidID = errors.New("invalid track ID")

// Lyrics orchestrates lyrics search and retrieval, applying input validation
// on top of a domain.LyricsClient.
type Lyrics struct {
	client domain.LyricsClient
}

// NewLyrics returns a Lyrics service backed by the given client.
func NewLyrics(client domain.LyricsClient) *Lyrics {
	return &Lyrics{client: client}
}

// Search trims and caps the query, then delegates to the client. A missing
// result (domain.ErrNotFound) is normalized to an empty slice.
func (s *Lyrics) Search(ctx context.Context, query string) ([]domain.Track, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []domain.Track{}, nil
	}
	if len(query) > maxQueryLength {
		query = query[:maxQueryLength]
	}
	tracks, err := s.client.Search(ctx, query)
	if errors.Is(err, domain.ErrNotFound) {
		return []domain.Track{}, nil
	}
	if err != nil {
		return nil, err
	}
	return tracks, nil
}

// GetByID validates the identifier and returns the matching track. It returns
// ErrInvalidID for non-positive IDs and propagates domain.ErrNotFound.
func (s *Lyrics) GetByID(ctx context.Context, id int) (*domain.Track, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}
	return s.client.GetByID(ctx, id)
}

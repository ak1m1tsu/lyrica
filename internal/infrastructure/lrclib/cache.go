package lrclib

import (
	"context"
	"time"

	"github.com/ak1m1tsu/lyrica/internal/domain"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

const (
	defaultCacheSize = 256
	defaultCacheTTL  = 10 * time.Minute
)

var _ domain.LyricsClient = (*CachingClient)(nil)

// CachingClient wraps a domain.LyricsClient and memoises successful responses.
// Errors are never cached. expirable.LRU is internally mutex-protected.
type CachingClient struct {
	inner   domain.LyricsClient
	search  *expirable.LRU[string, []domain.Track]
	getByID *expirable.LRU[int, domain.Track]
}

func NewCachingClient(inner domain.LyricsClient) *CachingClient {
	return &CachingClient{
		inner:   inner,
		search:  expirable.NewLRU[string, []domain.Track](defaultCacheSize, nil, defaultCacheTTL),
		getByID: expirable.NewLRU[int, domain.Track](defaultCacheSize, nil, defaultCacheTTL),
	}
}

func (c *CachingClient) Search(ctx context.Context, query string) ([]domain.Track, error) {
	if cached, ok := c.search.Get(query); ok {
		out := make([]domain.Track, len(cached))
		copy(out, cached)
		return out, nil
	}
	tracks, err := c.inner.Search(ctx, query)
	if err != nil {
		return nil, err
	}
	toStore := make([]domain.Track, len(tracks))
	copy(toStore, tracks)
	c.search.Add(query, toStore)
	return tracks, nil
}

func (c *CachingClient) GetByID(ctx context.Context, id int) (*domain.Track, error) {
	if cached, ok := c.getByID.Get(id); ok {
		t := cached
		return &t, nil
	}
	track, err := c.inner.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if track == nil {
		return nil, domain.ErrNotFound
	}
	c.getByID.Add(id, *track)
	return track, nil
}

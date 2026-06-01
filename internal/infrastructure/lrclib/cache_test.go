package lrclib

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ak1m1tsu/lrclib/internal/domain"
	"github.com/hashicorp/golang-lru/v2/expirable"
)

type mockLyricsClient struct {
	searchCalls  int
	getByIDCalls int
	searchFn     func(context.Context, string) ([]domain.Track, error)
	getByIDFn    func(context.Context, int) (*domain.Track, error)
}

func (m *mockLyricsClient) Search(ctx context.Context, query string) ([]domain.Track, error) {
	m.searchCalls++
	return m.searchFn(ctx, query)
}

func (m *mockLyricsClient) GetByID(ctx context.Context, id int) (*domain.Track, error) {
	m.getByIDCalls++
	return m.getByIDFn(ctx, id)
}

func newCachingClientWithTTL(inner domain.LyricsClient, size int, ttl time.Duration) *CachingClient {
	return &CachingClient{
		inner:   inner,
		search:  expirable.NewLRU[string, []domain.Track](size, nil, ttl),
		getByID: expirable.NewLRU[int, domain.Track](size, nil, ttl),
	}
}

var (
	sampleTrack  = domain.Track{ID: 1, TrackName: "Track One", ArtistName: "Artist"}
	sampleTracks = []domain.Track{sampleTrack, {ID: 2, TrackName: "Track Two"}}
)

func TestCachingClient_Search_Miss(t *testing.T) {
	mock := &mockLyricsClient{searchFn: func(_ context.Context, _ string) ([]domain.Track, error) {
		return sampleTracks, nil
	}}
	c := NewCachingClient(mock)

	got, err := c.Search(context.Background(), "rock")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(sampleTracks) {
		t.Fatalf("expected %d tracks, got %d", len(sampleTracks), len(got))
	}
	if mock.searchCalls != 1 {
		t.Fatalf("expected 1 inner call, got %d", mock.searchCalls)
	}
}

func TestCachingClient_Search_Hit(t *testing.T) {
	mock := &mockLyricsClient{searchFn: func(_ context.Context, _ string) ([]domain.Track, error) {
		return sampleTracks, nil
	}}
	c := NewCachingClient(mock)

	c.Search(context.Background(), "rock") //nolint:errcheck
	got, err := c.Search(context.Background(), "rock")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(sampleTracks) {
		t.Fatalf("expected %d tracks, got %d", len(sampleTracks), len(got))
	}
	if mock.searchCalls != 1 {
		t.Fatalf("expected 1 inner call on cache hit, got %d", mock.searchCalls)
	}
}

func TestCachingClient_Search_DifferentQueries(t *testing.T) {
	mock := &mockLyricsClient{searchFn: func(_ context.Context, _ string) ([]domain.Track, error) {
		return sampleTracks, nil
	}}
	c := NewCachingClient(mock)

	c.Search(context.Background(), "rock")  //nolint:errcheck
	c.Search(context.Background(), "metal") //nolint:errcheck

	if mock.searchCalls != 2 {
		t.Fatalf("expected 2 inner calls for distinct queries, got %d", mock.searchCalls)
	}
}

func TestCachingClient_Search_ErrorNotCached(t *testing.T) {
	errNetwork := errors.New("network error")
	calls := 0
	mock := &mockLyricsClient{searchFn: func(_ context.Context, _ string) ([]domain.Track, error) {
		calls++
		if calls == 1 {
			return nil, errNetwork
		}
		return sampleTracks, nil
	}}
	c := NewCachingClient(mock)

	if _, err := c.Search(context.Background(), "rock"); err == nil {
		t.Fatal("expected error on first call")
	}
	if _, err := c.Search(context.Background(), "rock"); err != nil {
		t.Fatalf("expected success on retry, got: %v", err)
	}
	if mock.searchCalls != 2 {
		t.Fatalf("expected 2 inner calls (error not cached), got %d", mock.searchCalls)
	}
}

func TestCachingClient_Search_EmptySliceIsCached(t *testing.T) {
	mock := &mockLyricsClient{searchFn: func(_ context.Context, _ string) ([]domain.Track, error) {
		return []domain.Track{}, nil
	}}
	c := NewCachingClient(mock)

	c.Search(context.Background(), "unknownxyz") //nolint:errcheck
	c.Search(context.Background(), "unknownxyz") //nolint:errcheck

	if mock.searchCalls != 1 {
		t.Fatalf("expected empty-slice result to be cached, got %d inner calls", mock.searchCalls)
	}
}

func TestCachingClient_Search_CallerMutationIsolated(t *testing.T) {
	mock := &mockLyricsClient{searchFn: func(_ context.Context, _ string) ([]domain.Track, error) {
		return []domain.Track{{ID: 1, TrackName: "Original"}}, nil
	}}
	c := NewCachingClient(mock)

	first, _ := c.Search(context.Background(), "q")
	first[0].TrackName = "Mutated"

	second, _ := c.Search(context.Background(), "q")
	if second[0].TrackName != "Original" {
		t.Fatalf("cache was corrupted by caller mutation: got %q", second[0].TrackName)
	}
}

func TestCachingClient_GetByID_Miss(t *testing.T) {
	mock := &mockLyricsClient{getByIDFn: func(_ context.Context, _ int) (*domain.Track, error) {
		t := sampleTrack
		return &t, nil
	}}
	c := NewCachingClient(mock)

	got, err := c.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != sampleTrack.ID {
		t.Fatalf("expected track ID %d, got %d", sampleTrack.ID, got.ID)
	}
	if mock.getByIDCalls != 1 {
		t.Fatalf("expected 1 inner call, got %d", mock.getByIDCalls)
	}
}

func TestCachingClient_GetByID_Hit(t *testing.T) {
	mock := &mockLyricsClient{getByIDFn: func(_ context.Context, _ int) (*domain.Track, error) {
		t := sampleTrack
		return &t, nil
	}}
	c := NewCachingClient(mock)

	c.GetByID(context.Background(), 1) //nolint:errcheck
	got, err := c.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != sampleTrack.ID {
		t.Fatalf("expected track ID %d, got %d", sampleTrack.ID, got.ID)
	}
	if mock.getByIDCalls != 1 {
		t.Fatalf("expected 1 inner call on cache hit, got %d", mock.getByIDCalls)
	}
}

func TestCachingClient_GetByID_ErrorNotCached(t *testing.T) {
	calls := 0
	mock := &mockLyricsClient{getByIDFn: func(_ context.Context, _ int) (*domain.Track, error) {
		calls++
		if calls == 1 {
			return nil, domain.ErrNotFound
		}
		t := sampleTrack
		return &t, nil
	}}
	c := NewCachingClient(mock)

	if _, err := c.GetByID(context.Background(), 99); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
	if _, err := c.GetByID(context.Background(), 99); err != nil {
		t.Fatalf("expected success on retry, got: %v", err)
	}
	if mock.getByIDCalls != 2 {
		t.Fatalf("expected 2 inner calls (error not cached), got %d", mock.getByIDCalls)
	}
}

func TestCachingClient_GetByID_PointerIsolated(t *testing.T) {
	mock := &mockLyricsClient{getByIDFn: func(_ context.Context, _ int) (*domain.Track, error) {
		tr := domain.Track{ID: 1, TrackName: "Original"}
		return &tr, nil
	}}
	c := NewCachingClient(mock)

	first, _ := c.GetByID(context.Background(), 1)
	first.TrackName = "Mutated"

	second, _ := c.GetByID(context.Background(), 1)
	if second.TrackName != "Original" {
		t.Fatalf("cache was corrupted by pointer mutation: got %q", second.TrackName)
	}
}

func TestCachingClient_GetByID_NilNotReturned(t *testing.T) {
	mock := &mockLyricsClient{getByIDFn: func(_ context.Context, _ int) (*domain.Track, error) {
		tr := sampleTrack
		return &tr, nil
	}}
	c := NewCachingClient(mock)

	c.GetByID(context.Background(), 1) //nolint:errcheck
	got, err := c.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == nil {
		t.Fatal("cache hit returned nil pointer")
	}
}

func TestCachingClient_TTL_Expiry(t *testing.T) {
	mock := &mockLyricsClient{searchFn: func(_ context.Context, _ string) ([]domain.Track, error) {
		return sampleTracks, nil
	}}
	c := newCachingClientWithTTL(mock, 256, 50*time.Millisecond)

	c.Search(context.Background(), "rock") //nolint:errcheck
	time.Sleep(100 * time.Millisecond)
	c.Search(context.Background(), "rock") //nolint:errcheck

	if mock.searchCalls != 2 {
		t.Fatalf("expected 2 inner calls after TTL expiry, got %d", mock.searchCalls)
	}
}

func TestCachingClient_LRU_Eviction(t *testing.T) {
	mock := &mockLyricsClient{searchFn: func(_ context.Context, _ string) ([]domain.Track, error) {
		return sampleTracks, nil
	}}
	c := newCachingClientWithTTL(mock, 2, 0) // size=2, no TTL

	// Fill cache: "a" is LRU, "b" is MRU after two adds.
	c.Search(context.Background(), "a") //nolint:errcheck
	c.Search(context.Background(), "b") //nolint:errcheck
	// Third add evicts LRU ("a"); cache now holds "b" and "c".
	c.Search(context.Background(), "c") //nolint:errcheck

	mock.searchCalls = 0

	// "b" is still cached — hit, no inner call.
	c.Search(context.Background(), "b") //nolint:errcheck
	if mock.searchCalls != 0 {
		t.Fatalf("expected cache hit for 'b', got %d inner calls", mock.searchCalls)
	}

	// "a" was evicted — miss, one inner call.
	c.Search(context.Background(), "a") //nolint:errcheck
	if mock.searchCalls != 1 {
		t.Fatalf("expected 1 inner call for evicted 'a', got %d", mock.searchCalls)
	}
}

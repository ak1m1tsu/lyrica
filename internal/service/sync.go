package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ak1m1tsu/lyrica/internal/domain"
	"github.com/ak1m1tsu/lyrica/internal/infrastructure/googledrive"
	"golang.org/x/oauth2"
)

// Sync orchestrates push/pull of favorites to/from Google Drive.
type Sync struct {
	store  domain.FavoritesStore
	client *googledrive.Client
}

// NewSync returns a Sync service backed by the given store and Drive client.
func NewSync(store domain.FavoritesStore, client *googledrive.Client) *Sync {
	return &Sync{store: store, client: client}
}

// Push exports local favorites as JSON and uploads to the Drive App Data folder.
// Persists any refreshed OAuth token and updates LastSyncAt on success.
func (s *Sync) Push(ctx context.Context) error {
	ts := s.buildTokenSource(ctx)
	tracks := s.store.GetAll()
	data, err := json.Marshal(tracks)
	if err != nil {
		return err
	}
	if err := s.client.Upload(ctx, ts, data); err != nil {
		return err
	}
	s.persistRefreshedToken(ctx, ts)
	return s.store.SetLastSyncAt(ctx, time.Now().UTC().Format(time.RFC3339))
}

// Pull downloads favorites from Drive and merges them into the local store.
// Merge strategy: any Drive track whose ID is not in the local set is added.
// Local tracks absent from Drive are never removed.
// Returns the number of tracks added and updates LastSyncAt on success.
func (s *Sync) Pull(ctx context.Context) (int, error) {
	ts := s.buildTokenSource(ctx)
	data, err := s.client.Download(ctx, ts)
	if err != nil {
		return 0, err
	}
	s.persistRefreshedToken(ctx, ts)
	if data == nil {
		return 0, nil
	}
	var remote []domain.Track
	if err := json.Unmarshal(data, &remote); err != nil {
		return 0, err
	}
	added := 0
	for _, t := range remote {
		if !s.store.Has(t.ID) {
			if err := s.store.Add(ctx, t); err != nil {
				return added, err
			}
			added++
		}
	}
	if err := s.store.SetLastSyncAt(ctx, time.Now().UTC().Format(time.RFC3339)); err != nil {
		return added, err
	}
	return added, nil
}

func (s *Sync) buildTokenSource(ctx context.Context) oauth2.TokenSource {
	return s.client.TokenSource(ctx, s.store.GetGoogleAccessToken(), s.store.GetGoogleRefreshToken())
}

// persistRefreshedToken reads the current token from the source (which may have
// been refreshed during the Drive API call) and persists it back to the store.
func (s *Sync) persistRefreshedToken(ctx context.Context, ts oauth2.TokenSource) {
	tok, err := ts.Token()
	if err != nil {
		return
	}
	_ = s.store.SetGoogleAccessToken(ctx, tok.AccessToken)
	if tok.RefreshToken != "" {
		_ = s.store.SetGoogleRefreshToken(ctx, tok.RefreshToken)
	}
}

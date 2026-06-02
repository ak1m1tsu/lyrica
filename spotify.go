package main

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const spotifyClientID = "4af08da5609f4123995ca2fb2642b522"

// Spotify allows http://127.0.0.1 redirect URIs for native/desktop apps (RFC 8252).
const spotifyRedirectURI = "http://127.0.0.1:27182/callback"

type spotifyService struct {
	mu      sync.Mutex
	stopCh  chan struct{}
	running bool
}

func newSpotifyService() *spotifyService {
	return &spotifyService{}
}

func (s *spotifyService) startPolling(
	ctx context.Context,
	auth *spotifyauth.Authenticator,
	token *oauth2.Token,
	onToken func(access, refresh string),
	onTrack func(trackName, artistName string),
) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	stopCh := make(chan struct{})
	s.stopCh = stopCh
	s.running = true
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
		}()

		httpClient := auth.Client(ctx, token)
		client := spotify.New(httpClient)
		lastID := spotify.ID("")
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stopCh:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				current, err := client.PlayerCurrentlyPlaying(ctx)
				if err != nil {
					slog.Warn("spotify: poll error", "error", err)
					continue
				}
				if current == nil || !current.Playing || current.Item == nil {
					continue
				}
				if current.Item.ID != lastID {
					lastID = current.Item.ID
					artistName := ""
					if len(current.Item.Artists) > 0 {
						artistName = current.Item.Artists[0].Name
					}
					if onTrack != nil {
						onTrack(current.Item.Name, artistName)
					}
				}
			}
		}
	}()
}

func (s *spotifyService) stopPolling() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running || s.stopCh == nil {
		return
	}
	close(s.stopCh)
	s.stopCh = nil
}

// Package googledrive provides a Google Drive client for OAuth2 authentication
// and App Data folder operations used by the sync feature.
package googledrive

import (
	"bytes"
	"context"
	"errors"
	"io"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// IsAuthError reports whether err is a Google API authentication/authorization
// failure (HTTP 401 or 403), indicating that the stored tokens are invalid or
// have been revoked.
func IsAuthError(err error) bool {
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) {
		return apiErr.Code == 401 || apiErr.Code == 403
	}
	return false
}

const (
	redirectURI       = "http://127.0.0.1:27183/callback"
	favoritesFileName = "lyrica-favorites.json"
)

// Client handles Google OAuth2 authorization and Drive App Data operations.
type Client struct {
	cfg *oauth2.Config
}

// New returns a Client using the provided OAuth2 credentials.
// Pass empty strings when credentials are unavailable; Configured() will return false.
func New(clientID, clientSecret string) *Client {
	return &Client{
		cfg: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  redirectURI,
			Scopes:       []string{drive.DriveAppdataScope},
		},
	}
}

// Configured reports whether OAuth2 credentials have been set.
func (c *Client) Configured() bool {
	return c.cfg.ClientID != "" && c.cfg.ClientSecret != ""
}

// RedirectURI returns the local callback URL used during the OAuth2 flow.
func (c *Client) RedirectURI() string {
	return redirectURI
}

// AuthURL returns the Google OAuth2 consent page URL.
func (c *Client) AuthURL(state, verifier string) string {
	return c.cfg.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
}

// Exchange trades an authorization code for OAuth2 tokens.
func (c *Client) Exchange(ctx context.Context, code, verifier string) (*oauth2.Token, error) {
	return c.cfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
}

// TokenSource returns an auto-refreshing token source from stored tokens.
// The token's Expiry is intentionally set to the past to force an immediate
// refresh, ensuring the access token is always valid on first use.
func (c *Client) TokenSource(ctx context.Context, accessToken, refreshToken string) oauth2.TokenSource {
	tok := &oauth2.Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
	}
	return c.cfg.TokenSource(ctx, tok)
}

// Upload writes data to lyrica-favorites.json in the Drive App Data folder.
// Creates the file if absent; patches it if it already exists.
func (c *Client) Upload(ctx context.Context, ts oauth2.TokenSource, data []byte) error {
	svc, err := drive.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return err
	}
	fileID, err := findFile(svc)
	if err != nil {
		return err
	}
	if fileID != "" {
		_, err = svc.Files.Update(fileID, nil).
			Media(bytes.NewReader(data)).
			Context(ctx).
			Do()
		return err
	}
	f := &drive.File{
		Name:    favoritesFileName,
		Parents: []string{"appDataFolder"},
	}
	_, err = svc.Files.Create(f).
		Media(bytes.NewReader(data)).
		Context(ctx).
		Do()
	return err
}

// Download fetches lyrica-favorites.json from the Drive App Data folder.
// Returns nil, nil if no file exists yet (first sync from this device).
func (c *Client) Download(ctx context.Context, ts oauth2.TokenSource) ([]byte, error) {
	svc, err := drive.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, err
	}
	fileID, err := findFile(svc)
	if err != nil {
		return nil, err
	}
	if fileID == "" {
		return nil, nil
	}
	resp, err := svc.Files.Get(fileID).Context(ctx).Download()
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func findFile(svc *drive.Service) (string, error) {
	list, err := svc.Files.List().
		Spaces("appDataFolder").
		Q("name='" + favoritesFileName + "'").
		Fields("files(id)").
		Do()
	if err != nil {
		return "", err
	}
	if len(list.Files) == 0 {
		return "", nil
	}
	return list.Files[0].Id, nil
}

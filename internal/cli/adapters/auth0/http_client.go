package auth0

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"exercise/internal/cli/domain/auth"

	"golang.org/x/oauth2"
)

func NewHTTPClient(cfg Config, store auth.TokenStore) *http.Client {
	return &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.ReuseTokenSource(nil, &persistingTokenSource{cfg: cfg, store: store}),
			Base:   cfg.Transport,
		},
		Timeout: httpTimeout,
	}
}

// Refresh tokens rotate: each refresh invalidates the previous one, so every new token must be
// saved or the next run fails with invalid_grant. The store is read on the first request (not at
// construction) so the client can be wired at startup, before the user has logged in.
type persistingTokenSource struct {
	cfg   Config
	store auth.TokenStore
	src   oauth2.TokenSource
	last  *oauth2.Token
}

func (s *persistingTokenSource) Token() (*oauth2.Token, error) {
	if s.src == nil {
		tok, err := s.store.Load()
		if err != nil {
			return nil, fmt.Errorf("load token: %w", err)
		}
		// Token() has no ctx; this one only carries the refresh HTTP client (bounded by httpTimeout).
		s.src = oauthConfig(s.cfg).TokenSource(withHTTPClient(context.Background(), s.cfg), tok)
		s.last = tok
	}

	tok, err := s.src.Token()
	if err != nil {
		return nil, s.describeError(err)
	}
	if tok.AccessToken != s.last.AccessToken {
		if err := s.store.Save(tok); err != nil {
			return nil, fmt.Errorf("save refreshed token: %w", err)
		}
		s.last = tok
	}
	return tok, nil
}

func (s *persistingTokenSource) describeError(err error) error {
	// No refresh token: x/oauth2 can't refresh and returns a plain error.
	if s.last.RefreshToken == "" {
		return auth.ErrSessionExpired
	}
	// invalid_grant: refresh token expired (idle/absolute lifetime), revoked, or already rotated.
	if re, ok := errors.AsType[*oauth2.RetrieveError](err); ok && re.ErrorCode == "invalid_grant" {
		return auth.ErrSessionExpired
	}
	return fmt.Errorf("refresh access token: %w", err)
}

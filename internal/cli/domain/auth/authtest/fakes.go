package authtest

import (
	"context"
	"sync"

	"exercise/internal/cli/domain/auth"

	"golang.org/x/oauth2"
)

type FakeAuthenticator struct {
	// Exported, unlike the repository fakes: outcomes (denied, expired...) come from the IdP,
	// not from stored state, so each test sets the one it needs.
	Code  auth.DeviceCode
	Token *oauth2.Token
	Err   error
}

func (f *FakeAuthenticator) DeviceLogin(ctx context.Context, show func(auth.DeviceCode)) (*oauth2.Token, error) {
	if ctx.Err() != nil {
		return nil, auth.ErrLoginCancelled
	}
	// Before failing, like the real flow: denial and expiry happen while polling.
	show(f.Code)
	if f.Err != nil {
		return nil, f.Err
	}
	return f.Token, nil
}

type FakeTokenStore struct {
	mu    sync.Mutex
	saved *oauth2.Token
	Err   error
}

func (f *FakeTokenStore) Load() (*oauth2.Token, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.saved == nil {
		return nil, auth.ErrNotLoggedIn
	}
	return f.saved, nil
}

func (f *FakeTokenStore) Save(tok *oauth2.Token) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.Err != nil {
		return f.Err
	}
	f.saved = tok
	return nil
}

func (f *FakeTokenStore) Path() string { return "/fake/token.json" }

func (f *FakeTokenStore) Saved() *oauth2.Token {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.saved
}

package auth

import (
	"context"
	"errors"

	"golang.org/x/oauth2"
)

var (
	ErrLoginCancelled = errors.New("cancelled")
	ErrLoginDenied    = errors.New("denied in the browser")
	ErrLoginExpired   = errors.New("code expired; run `exercise login` again")
	ErrNotLoggedIn    = errors.New("not logged in; run `exercise login`")
	ErrSessionExpired = errors.New("session expired; run `exercise login`")
)

type DeviceCode struct {
	// Has the user code pre-filled, so the user only has to confirm it.
	VerificationURL string
	UserCode        string
}

type Authenticator interface {
	// Returns oauth2.Token as-is: a plain value type the token source and store already speak.
	// show is a callback, not printing: each frontend presents the code its own way (terminal, MCP...).
	DeviceLogin(ctx context.Context, show func(DeviceCode)) (*oauth2.Token, error)
}

type TokenStore interface {
	// ErrNotLoggedIn if nothing was saved yet.
	Load() (*oauth2.Token, error)
	Save(tok *oauth2.Token) error
	Path() string
}

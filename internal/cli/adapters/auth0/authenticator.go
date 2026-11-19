package auth0

import (
	"context"
	"errors"
	"fmt"
	"time"

	"exercise/internal/cli/domain/auth"

	"golang.org/x/oauth2"
)

type Authenticator struct {
	cfg Config
}

func NewAuthenticator(cfg Config) *Authenticator {
	return &Authenticator{cfg: cfg}
}

func (a *Authenticator) DeviceLogin(ctx context.Context, show func(auth.DeviceCode)) (*oauth2.Token, error) {
	oc := oauthConfig(a.cfg)
	ctx = withHTTPClient(ctx, a.cfg)

	// audience makes Auth0 issue a JWT access token for our API instead of an opaque one.
	da, err := oc.DeviceAuth(ctx, oauth2.SetAuthURLParam("audience", a.cfg.Audience))
	if err != nil {
		return nil, fmt.Errorf("request device code: %w", err)
	}

	show(auth.DeviceCode{VerificationURL: da.VerificationURIComplete, UserCode: da.UserCode})

	tok, err := oc.DeviceAccessToken(ctx, da)
	if err != nil {
		return nil, describeDeviceError(err, da.Expiry)
	}
	return tok, nil
}

func describeDeviceError(err error, expiry time.Time) error {
	if errors.Is(err, context.Canceled) {
		return auth.ErrLoginCancelled
	}
	// DeviceAccessToken bounds ctx with WithDeadline(expiry), so a code that expires
	// locally surfaces as DeadlineExceeded rather than expired_token. Check the clock too:
	// a single poll hitting httpTimeout also matches DeadlineExceeded (net/http timeoutError).
	if errors.Is(err, context.DeadlineExceeded) && !expiry.IsZero() && !time.Now().Before(expiry) {
		return auth.ErrLoginExpired
	}
	if re, ok := errors.AsType[*oauth2.RetrieveError](err); ok {
		switch re.ErrorCode {
		case "access_denied":
			return auth.ErrLoginDenied
		case "expired_token":
			return auth.ErrLoginExpired
		}
	}
	return fmt.Errorf("wait for device approval: %w", err)
}

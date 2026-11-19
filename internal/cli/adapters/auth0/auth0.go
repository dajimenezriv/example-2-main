package auth0

import (
	"context"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// Per request, not total: login polling runs until the device code expires.
const httpTimeout = 10 * time.Second

type Config struct {
	Issuer   string
	Audience string
	ClientID string
	// nil: http.DefaultTransport. Tests set an in-memory one: synctest's fake clock can't run with
	// real sockets, and device-flow polling waits seconds between requests.
	Transport http.RoundTripper
}

func oauthConfig(cfg Config) *oauth2.Config {
	issuer := strings.TrimSuffix(cfg.Issuer, "/")
	return &oauth2.Config{
		ClientID: cfg.ClientID,
		Endpoint: oauth2.Endpoint{
			DeviceAuthURL: issuer + "/oauth/device/code",
			TokenURL:      issuer + "/oauth/token",
			// Public client: client_id goes in the body, no secret.
			AuthStyle: oauth2.AuthStyleInParams,
		},
		// offline_access asks for a refresh token.
		Scopes: []string{"openid", "profile", "email", "offline_access"},
	}
}

// x/oauth2 takes its HTTP client from the context (device code, polling, refresh). Without it,
// it falls back to http.DefaultClient, which has no timeout.
func withHTTPClient(ctx context.Context, cfg Config) context.Context {
	return context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Timeout: httpTimeout, Transport: cfg.Transport})
}

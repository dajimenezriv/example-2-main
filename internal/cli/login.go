package cli

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// httpTimeout bounds each request to the IdP, including every poll.
const httpTimeout = 10 * time.Second

var (
	ErrLoginCancelled = errors.New("login cancelled")
	ErrLoginDenied    = errors.New("login denied in the browser")
	ErrLoginExpired   = errors.New("login code expired; run login again")
)

func newLoginCmd(cfg *Config) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Log in through the browser (OAuth2 Device Authorization flow)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := DefaultTokenStore()
			if err != nil {
				return err
			}

			tok, err := deviceLogin(cmd, cfg)
			if err != nil {
				return err
			}
			if err := store.Save(tok); err != nil {
				return err
			}

			// Status messages go to stderr (cmd.PrintErr*), keeping stdout for data output.
			cmd.PrintErrf("Logged in. Tokens saved to %s\n", store.Path())
			if tok.RefreshToken == "" {
				cmd.PrintErrln("Warning: no refresh token issued; you will need to log in again when the access token expires.")
			}
			return nil
		},
	}
}

// deviceLogin runs RFC 8628: request a device + user code, show the user where to approve,
// then poll the token endpoint until they approve, deny, or the code expires.
func deviceLogin(cmd *cobra.Command, cfg *Config) (*oauth2.Token, error) {
	issuer := strings.TrimSuffix(cfg.Issuer, "/")
	oc := &oauth2.Config{
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

	// x/oauth2 picks up the HTTP client from the context.
	ctx := context.WithValue(cmd.Context(), oauth2.HTTPClient, &http.Client{Timeout: httpTimeout})

	// audience makes Auth0 issue a JWT access token for our API instead of an opaque one.
	da, err := oc.DeviceAuth(ctx, oauth2.SetAuthURLParam("audience", cfg.Audience))
	if err != nil {
		return nil, fmt.Errorf("request device code: %w", err)
	}

	cmd.PrintErrf("Open this URL in a browser:\n\n  %s\n\n", verificationURL(da))
	cmd.PrintErrf("Confirm the code shown there is: %s\n\nWaiting for approval (Ctrl+C to cancel)...\n", da.UserCode)

	// Polls at da.Interval, backs off on slow_down, stops on ctx cancel or da.Expiry.
	tok, err := oc.DeviceAccessToken(ctx, da)
	if err != nil {
		return nil, describeDeviceError(err)
	}
	return tok, nil
}

// verificationURL prefers the URL with the code pre-filled, so the user only has to confirm it.
func verificationURL(da *oauth2.DeviceAuthResponse) string {
	if da.VerificationURIComplete != "" {
		return da.VerificationURIComplete
	}
	return da.VerificationURI
}

func describeDeviceError(err error) error {
	if errors.Is(err, context.Canceled) {
		return ErrLoginCancelled
	}
	if re, ok := errors.AsType[*oauth2.RetrieveError](err); ok {
		switch re.ErrorCode {
		case "access_denied":
			return ErrLoginDenied
		case "expired_token":
			return ErrLoginExpired
		}
	}
	return fmt.Errorf("wait for device approval: %w", err)
}

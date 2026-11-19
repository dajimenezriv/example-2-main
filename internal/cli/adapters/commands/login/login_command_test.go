package login

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"exercise/internal/cli/adapters/commands/cmdtest"
	"exercise/internal/cli/domain/auth"
	"exercise/internal/cli/domain/auth/authtest"

	"golang.org/x/oauth2"
)

var (
	deviceCode = auth.DeviceCode{
		VerificationURL: "https://idp.example/activate?user_code=ABCD-EFGH",
		UserCode:        "ABCD-EFGH",
		CodePrefilled:   true,
	}
	errDiskFull = errors.New("disk full")
)

const prompt = "Open this URL in a browser:\n\n" +
	"  https://idp.example/activate?user_code=ABCD-EFGH\n\n" +
	"Confirm the code shown there is: ABCD-EFGH\n\n" +
	"Waiting for approval (Ctrl+C to cancel)...\n"

func login(ctx context.Context, t *testing.T, a auth.Authenticator, s auth.TokenStore, args ...string) cmdtest.Result {
	t.Helper()
	return cmdtest.Run(ctx, t, NewCommand(a, s).Register, append([]string{"login"}, args...)...)
}

func TestLogin_Success(t *testing.T) {
	tests := []struct {
		name       string
		token      *oauth2.Token
		wantStderr string
	}{
		{
			name:       "with refresh token",
			token:      &oauth2.Token{AccessToken: "access", RefreshToken: "refresh"},
			wantStderr: prompt + "Logged in. Tokens saved to /fake/token.json\n",
		},
		{
			name:  "without refresh token warns that the session won't renew",
			token: &oauth2.Token{AccessToken: "access"},
			wantStderr: prompt + "Logged in. Tokens saved to /fake/token.json\n" +
				"Warning: no refresh token issued; you will need to log in again when the access token expires.\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &authtest.FakeTokenStore{}

			res := login(t.Context(), t, &authtest.FakeAuthenticator{Code: deviceCode, Token: tt.token}, store)

			require.NoError(t, res.Err)
			assert.Equal(t, tt.token, store.Saved())
			assert.Equal(t, tt.wantStderr, res.Stderr)
			// Everything is a status message: nothing for scripts on stdout.
			assert.Empty(t, res.Stdout)
		})
	}
}

func TestLogin_CodeNotPrefilledAsksToEnterIt(t *testing.T) {
	code := auth.DeviceCode{VerificationURL: "https://idp.example/activate", UserCode: "ABCD-EFGH"}
	token := &oauth2.Token{AccessToken: "access", RefreshToken: "refresh"}

	res := login(t.Context(), t, &authtest.FakeAuthenticator{Code: code, Token: token}, &authtest.FakeTokenStore{})

	require.NoError(t, res.Err)
	assert.Equal(t, "Open this URL in a browser:\n\n"+
		"  https://idp.example/activate\n\n"+
		"Enter this code there: ABCD-EFGH\n\n"+
		"Waiting for approval (Ctrl+C to cancel)...\n"+
		"Logged in. Tokens saved to /fake/token.json\n", res.Stderr)
}

func TestLogin_Errors(t *testing.T) {
	canceled, cancel := context.WithCancel(t.Context())
	cancel()
	token := &oauth2.Token{AccessToken: "access", RefreshToken: "refresh"}

	tests := []struct {
		name       string
		ctx        context.Context
		authErr    error
		storeErr   error
		wantErr    error
		wantMsg    string
		wantStderr string
	}{
		{
			name:       "denied in the browser",
			ctx:        t.Context(),
			authErr:    auth.ErrLoginDenied,
			wantErr:    auth.ErrLoginDenied,
			wantMsg:    "login: auth: denied in the browser",
			wantStderr: prompt,
		},
		{
			name:       "code expired",
			ctx:        t.Context(),
			authErr:    auth.ErrLoginExpired,
			wantErr:    auth.ErrLoginExpired,
			wantMsg:    "login: auth: code expired; run `exercise login` again",
			wantStderr: prompt,
		},
		{
			name:    "cancelled before the code is shown",
			ctx:     canceled,
			wantErr: auth.ErrLoginCancelled,
			wantMsg: "login: auth: cancelled",
		},
		{
			name:       "token approved but not saved",
			ctx:        t.Context(),
			storeErr:   errDiskFull,
			wantErr:    errDiskFull,
			wantMsg:    "login: save token: disk full",
			wantStderr: prompt,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &authtest.FakeTokenStore{Err: tt.storeErr}
			authenticator := &authtest.FakeAuthenticator{Code: deviceCode, Token: token, Err: tt.authErr}

			res := login(tt.ctx, t, authenticator, store)

			require.ErrorIs(t, res.Err, tt.wantErr)
			require.EqualError(t, res.Err, tt.wantMsg)
			assert.Nil(t, store.Saved())
			// No "Logged in" line on any failure.
			assert.Equal(t, tt.wantStderr, res.Stderr)
			assert.Empty(t, res.Stdout)
		})
	}
}

func TestLogin_RejectsArguments(t *testing.T) {
	store := &authtest.FakeTokenStore{}
	authenticator := &authtest.FakeAuthenticator{Code: deviceCode, Token: &oauth2.Token{AccessToken: "access"}}

	res := login(t.Context(), t, authenticator, store, "extra")

	require.Error(t, res.Err)
	assert.Contains(t, res.Err.Error(), `unknown command "extra" for "exercise login"`)
	assert.Nil(t, store.Saved())
}

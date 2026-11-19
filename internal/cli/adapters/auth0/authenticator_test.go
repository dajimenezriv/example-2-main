package auth0

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"exercise/internal/cli/domain/auth"
)

// x/oauth2's DeviceAuth never closes the device-code response body (deviceauth.go, v0.38), so the
// goroutine net/http starts to enforce Client.Timeout lingers until httpTimeout fires. Let fake time
// run past it before the bubble ends, or synctest reports it as a deadlock.
func waitOutLeakedDeviceAuthTimer() {
	time.Sleep(httpTimeout)
}

var wantCode = auth.DeviceCode{
	VerificationURL: "https://tenant.example/activate?user_code=ABCD-EFGH",
	UserCode:        "ABCD-EFGH",
}

func TestDeviceLogin_Success(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		defer waitOutLeakedDeviceAuthTimer()
		start := time.Now()
		f := &fakeAuth0{
			device: deviceReply(900),
			tokens: []reply{oauthError("authorization_pending"), oauthError("authorization_pending"), tokenReply("access-1", "refresh-1")},
		}
		var shown []auth.DeviceCode

		tok, err := NewAuthenticator(newConfig(f)).DeviceLogin(t.Context(), func(dc auth.DeviceCode) { shown = append(shown, dc) })

		require.NoError(t, err)
		assert.Equal(t, "access-1", tok.AccessToken)
		assert.Equal(t, "refresh-1", tok.RefreshToken)
		assert.Equal(t, []auth.DeviceCode{wantCode}, shown)

		require.Len(t, f.deviceForms, 1)
		assert.Equal(t, "cli-client", f.deviceForms[0].Get("client_id"))
		assert.Equal(t, "https://participants-api", f.deviceForms[0].Get("audience"))
		assert.Equal(t, "openid profile email offline_access", f.deviceForms[0].Get("scope"))

		// Polls every `interval` (5s) of fake time: the test still runs instantly.
		assert.Equal(t, []time.Duration{5 * time.Second, 10 * time.Second, 15 * time.Second}, since(start, f.tokenTimes))
		for _, form := range f.tokenForms {
			assert.Equal(t, "urn:ietf:params:oauth:grant-type:device_code", form.Get("grant_type"))
			assert.Equal(t, "dev-123", form.Get("device_code"))
			assert.Equal(t, "cli-client", form.Get("client_id"))
			// Public client: a secret would mean a misconfigured app type.
			assert.False(t, form.Has("client_secret"))
		}
	})
}

func TestDeviceLogin_SlowDownAddsFiveSeconds(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		defer waitOutLeakedDeviceAuthTimer()
		start := time.Now()
		f := &fakeAuth0{
			device: deviceReply(900),
			tokens: []reply{oauthError("slow_down"), oauthError("authorization_pending"), tokenReply("access-1", "refresh-1")},
		}

		_, err := NewAuthenticator(newConfig(f)).DeviceLogin(t.Context(), func(auth.DeviceCode) {})

		require.NoError(t, err)
		assert.Equal(t, []time.Duration{5 * time.Second, 15 * time.Second, 25 * time.Second}, since(start, f.tokenTimes))
	})
}

func TestDeviceLogin_Failures(t *testing.T) {
	tests := []struct {
		name        string
		device      reply
		tokens      []reply
		wantErrIs   error
		wantMessage string
		wantShown   bool
		wantPolls   []time.Duration
	}{
		{
			name:      "denied in the browser",
			device:    deviceReply(900),
			tokens:    []reply{oauthError("authorization_pending"), oauthError("access_denied")},
			wantErrIs: auth.ErrLoginDenied,
			wantShown: true,
			wantPolls: []time.Duration{5 * time.Second, 10 * time.Second},
		},
		{
			name:      "IdP reports the code expired",
			device:    deviceReply(900),
			tokens:    []reply{oauthError("expired_token")},
			wantErrIs: auth.ErrLoginExpired,
			wantShown: true,
			wantPolls: []time.Duration{5 * time.Second},
		},
		{
			// x/oauth2 stops at the code's expiry (a local deadline) before the IdP says so:
			// this surfaces as DeadlineExceeded, which must still read as "expired".
			name:      "code expires locally while still pending",
			device:    deviceReply(12),
			tokens:    []reply{oauthError("authorization_pending")},
			wantErrIs: auth.ErrLoginExpired,
			wantShown: true,
			wantPolls: []time.Duration{5 * time.Second, 10 * time.Second},
		},
		{
			name:        "unexpected polling error is wrapped, not mapped",
			device:      deviceReply(900),
			tokens:      []reply{oauthError("server_error")},
			wantMessage: "wait for device approval",
			wantShown:   true,
			wantPolls:   []time.Duration{5 * time.Second},
		},
		{
			name:        "device code request rejected (e.g. wrong client ID)",
			device:      reply{status: 401, body: `{"error": "unauthorized_client"}`},
			wantMessage: "request device code",
			wantPolls:   []time.Duration{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				defer waitOutLeakedDeviceAuthTimer()
				start := time.Now()
				f := &fakeAuth0{device: tt.device, tokens: tt.tokens}
				shown := false

				tok, err := NewAuthenticator(newConfig(f)).DeviceLogin(t.Context(), func(auth.DeviceCode) { shown = true })

				require.Error(t, err)
				assert.Nil(t, tok)
				if tt.wantErrIs != nil {
					require.ErrorIs(t, err, tt.wantErrIs)
				} else {
					assert.Contains(t, err.Error(), tt.wantMessage)
				}
				assert.Equal(t, tt.wantShown, shown)
				assert.Equal(t, tt.wantPolls, since(start, f.tokenTimes))
			})
		})
	}
}

func TestDeviceLogin_CancelledWhilePolling(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		defer waitOutLeakedDeviceAuthTimer()
		start := time.Now()
		f := &fakeAuth0{device: deviceReply(900), tokens: []reply{oauthError("authorization_pending")}}
		ctx, cancel := context.WithCancel(t.Context())
		go func() {
			time.Sleep(7 * time.Second) // Ctrl+C between the first and second poll.
			cancel()
		}()

		_, err := NewAuthenticator(newConfig(f)).DeviceLogin(ctx, func(auth.DeviceCode) {})

		require.ErrorIs(t, err, auth.ErrLoginCancelled)
		assert.Equal(t, []time.Duration{5 * time.Second}, since(start, f.tokenTimes))
		assert.Equal(t, 7*time.Second, time.Since(start))
	})
}

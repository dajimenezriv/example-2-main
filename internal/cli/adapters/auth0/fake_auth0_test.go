package auth0

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"time"
)

type reply struct {
	status int
	body   string
}

func oauthError(code string) reply {
	return reply{status: http.StatusForbidden, body: fmt.Sprintf(`{"error": %q}`, code)}
}

func tokenReply(access, refresh string) reply {
	body := fmt.Sprintf(`{"access_token": %q, "token_type": "Bearer", "expires_in": 3600`, access)
	if refresh != "" {
		body += fmt.Sprintf(`, "refresh_token": %q`, refresh)
	}
	return reply{status: http.StatusOK, body: body + "}"}
}

func deviceReply(expiresIn int) reply {
	return deviceReplyWith(expiresIn, `"verification_uri_complete": "https://tenant.example/activate?user_code=ABCD-EFGH",`)
}

// Optional per RFC 8628: some IdPs only send verification_uri.
func deviceReplyWithoutCompleteURI(expiresIn int) reply {
	return deviceReplyWith(expiresIn, "")
}

func deviceReplyWith(expiresIn int, extra string) reply {
	return reply{status: http.StatusOK, body: fmt.Sprintf(`{
		"device_code": "dev-123",
		"user_code": "ABCD-EFGH",
		"verification_uri": "https://tenant.example/activate",
		%s
		"expires_in": %d,
		"interval": 5
	}`, extra, expiresIn)}
}

// fakeAuth0 is an http.RoundTripper, not an httptest.Server: synctest's fake clock only advances
// when every goroutine is blocked on something inside the bubble, which real sockets are not.
type fakeAuth0 struct {
	mu sync.Mutex

	device reply
	// Served in order; the last one repeats (e.g. authorization_pending until the code expires).
	tokens []reply

	deviceForms []url.Values
	tokenForms  []url.Values
	tokenTimes  []time.Time
	apiAuth     []string
}

func (f *fakeAuth0) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Body != nil { // nil for GETs.
		defer r.Body.Close() //nolint:errcheck // RoundTripper must close the body; nothing to do on failure.
	}
	rec := httptest.NewRecorder()

	f.mu.Lock()
	defer f.mu.Unlock()

	switch r.URL.Path {
	case "/oauth/device/code":
		f.deviceForms = append(f.deviceForms, parseForm(r))
		write(rec, f.device)
	case "/oauth/token":
		f.tokenForms = append(f.tokenForms, parseForm(r))
		f.tokenTimes = append(f.tokenTimes, time.Now())
		write(rec, f.tokens[min(len(f.tokenForms), len(f.tokens))-1])
	case "/actors":
		f.apiAuth = append(f.apiAuth, r.Header.Get("Authorization"))
		write(rec, reply{status: http.StatusOK, body: "{}"})
	default:
		write(rec, reply{status: http.StatusNotFound, body: "{}"})
	}
	return rec.Result(), nil
}

func parseForm(r *http.Request) url.Values {
	if err := r.ParseForm(); err != nil {
		panic(err)
	}
	return r.PostForm
}

func write(w http.ResponseWriter, rep reply) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(rep.status)
	_, _ = w.Write([]byte(rep.body)) //nolint:errcheck // httptest.ResponseRecorder never fails.
}

func newConfig(f *fakeAuth0) Config {
	return Config{
		Issuer:    "https://tenant.example/",
		Audience:  "https://participants-api",
		ClientID:  "cli-client",
		Transport: f,
	}
}

// Offsets from the start of the bubble's fake clock (midnight UTC 2000-01-01).
func since(start time.Time, times []time.Time) []time.Duration {
	out := make([]time.Duration, len(times))
	for i, ts := range times {
		out[i] = ts.Sub(start)
	}
	return out
}

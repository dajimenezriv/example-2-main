package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"exercise/internal/cli/domain/auth"
)

var ErrUnexpectedResponse = errors.New("unexpected API response")

type Error struct {
	// Huma's RFC 9457 problem+json body.
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
	Errors []struct {
		Message  string `json:"message"`
		Location string `json:"location"`
	} `json:"errors"`
}

func (e *Error) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "API %d %s", e.Status, e.Title)
	if e.Detail != "" {
		b.WriteString(": ")
		b.WriteString(e.Detail)
	}
	// Validation failures (422) list each invalid field.
	for _, fe := range e.Errors {
		fmt.Fprintf(&b, "\n  - %s: %s", fe.Location, fe.Message)
	}
	return b.String()
}

// Parses the raw body instead of the generated ApplicationproblemJSON<status> fields:
// one code path for every operation and status, documented or not.
func newAPIError(status int, body []byte) *Error {
	apiErr := &Error{}
	if err := json.Unmarshal(body, apiErr); err != nil {
		// Not problem+json (e.g. a proxy's HTML page): show the raw body.
		apiErr = &Error{Detail: strings.TrimSpace(string(body))}
	}
	apiErr.Status = status
	if apiErr.Title == "" {
		apiErr.Title = http.StatusText(status)
	}
	return apiErr
}

// Token errors arrive wrapped in *url.Error; they are already actionable without the
// `Get "http://...":` noise.
func describeTransportError(err error) error {
	for _, tokenErr := range []error{auth.ErrNotLoggedIn, auth.ErrSessionExpired} {
		if errors.Is(err, tokenErr) {
			return tokenErr
		}
	}
	return fmt.Errorf("call API: %w", err)
}

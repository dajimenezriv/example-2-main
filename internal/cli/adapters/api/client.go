package api

import (
	"fmt"
	"net/http"

	"exercise/internal/cli/adapters/api/apiclient"
)

func NewClient(baseURL string, httpClient *http.Client) (*apiclient.ClientWithResponses, error) {
	c, err := apiclient.NewClientWithResponses(baseURL, apiclient.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("create API client: %w", err)
	}
	return c, nil
}

// Satisfied by every generated <Op>HTTPResponse, so one helper covers all operations.
type httpResponse interface {
	StatusCode() int
	GetBody() []byte
}

func checkResponse[R httpResponse](resp R, err error) (R, error) {
	if err != nil {
		return resp, describeTransportError(err)
	}
	if code := resp.StatusCode(); code < http.StatusOK || code >= http.StatusMultipleChoices {
		return resp, newAPIError(code, resp.GetBody())
	}
	return resp, nil
}

// The generated JSON<status> field is nil if a 2xx response wasn't the documented JSON
// (e.g. a proxy answered instead of the API).
func requireBody[T any](body *T) (*T, error) {
	if body == nil {
		return nil, ErrUnexpectedResponse
	}
	return body, nil
}

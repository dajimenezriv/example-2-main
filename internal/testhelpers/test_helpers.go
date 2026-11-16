package testhelpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"exercise/internal/platform/httpapi"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func NewHumaAPI(t *testing.T) humatest.TestAPI {
	t.Helper()
	_, api := humatest.New(t, httpapi.Config())
	return api
}

func AssertHumaOutput(t *testing.T, rec *httptest.ResponseRecorder, wantCode int, wantBody string) {
	t.Helper()

	require.Equal(t, wantCode, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

	got := rec.Body.String()

	var g, w any
	err := json.Unmarshal([]byte(got), &g)
	require.NoError(t, err)
	err = json.Unmarshal([]byte(wantBody), &w)
	require.NoError(t, err)
	assert.Equal(t, w, g)
}

func AssertHumaError(t *testing.T, rec *httptest.ResponseRecorder, wantCode int, wantDetail string, wantErrors ...*huma.ErrorDetail) {
	t.Helper()

	require.Equal(t, wantCode, rec.Code)
	assert.Equal(t, "application/problem+json", rec.Header().Get("Content-Type"))

	var got huma.ErrorModel
	err := json.Unmarshal(rec.Body.Bytes(), &got)
	require.NoError(t, err)

	want := huma.ErrorModel{
		Status: wantCode,
		Title:  http.StatusText(wantCode),
		Detail: wantDetail,
		Errors: wantErrors,
	}
	assert.Equal(t, want, got)
}

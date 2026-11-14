package testhelpers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"exercise/internal/platform/httpapi"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/humatest"
	"github.com/google/go-cmp/cmp"
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

	if rec.Code != wantCode {
		t.Errorf("Code = %d, want %d", rec.Code, wantCode)
	}
	if got, want := rec.Header().Get("Content-Type"), "application/json"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}

	got := rec.Body.String()

	var g, w any
	_ = json.Unmarshal([]byte(got), &g)
	_ = json.Unmarshal([]byte(wantBody), &w)
	if diff := cmp.Diff(w, g); diff != "" {
		t.Errorf("JSON mismatch (-want +got):\n%s", diff)
	}
}

func AssertHumaError(t *testing.T, rec *httptest.ResponseRecorder, wantCode int, wantDetail string, wantErrors ...*huma.ErrorDetail) {
	t.Helper()

	assert.Equal(t, wantCode, rec.Code)
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

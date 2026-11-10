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
)

func NewHumaAPI(t *testing.T) humatest.TestAPI {
	t.Helper()
	_, api := humatest.New(t, httpapi.Config())
	return api
}

func AssertHumaOutput(t testing.TB, rec *httptest.ResponseRecorder, wantCode int, want string) {
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
	_ = json.Unmarshal([]byte(want), &w)
	if diff := cmp.Diff(w, g); diff != "" {
		t.Errorf("JSON mismatch (-want +got):\n%s", diff)
	}
}

func AssertHumaError(t testing.TB, rec *httptest.ResponseRecorder, wantCode int, wantDetail string, wantErrors ...*huma.ErrorDetail) {
	t.Helper()

	if got := rec.Code; got != wantCode {
		t.Errorf("Code = %d, want %d", got, wantCode)
	}
	if got, want := rec.Header().Get("Content-Type"), "application/problem+json"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}

	var got huma.ErrorModel
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	want := huma.ErrorModel{
		Status: wantCode,
		Title:  http.StatusText(wantCode),
		Detail: wantDetail,
		Errors: wantErrors,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("error body mismatch (-want +got):\n%s", diff)
	}
}

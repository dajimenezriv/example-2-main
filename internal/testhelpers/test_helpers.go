package testhelpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"exercise/internal/platform/httpapi"

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
	json.Unmarshal([]byte(got), &g)
	json.Unmarshal([]byte(want), &w)
	if diff := cmp.Diff(w, g); diff != "" {
		t.Errorf("JSON mismatch (-want +got):\n%s", diff)
	}
}

func AssertHumaError(t testing.TB, rec *httptest.ResponseRecorder, wantCode int, wantDetail string) {
	t.Helper()

	if got := rec.Code; got != wantCode {
		t.Errorf("Code = %d, want %d", got, wantCode)
	}
	if got, want := rec.Header().Get("Content-Type"), "application/problem+json"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}

	var g, w any
	got := rec.Body.String()
	want := fmt.Sprintf(`{"status": %d, "title": %q, "detail": %q}`,
		wantCode,
		http.StatusText(wantCode),
		wantDetail)

	json.Unmarshal([]byte(got), &g)
	json.Unmarshal([]byte(want), &w)
	if diff := cmp.Diff(w, g); diff != "" {
		t.Errorf("JSON mismatch (-want +got):\n%s", diff)
	}
}

package chiapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dajimenezriv/personal/internal/db"
)

type fakeLinksStore struct {
	create func(ctx context.Context, url, code string) (db.Link, error)
	get    func(ctx context.Context, code string) (db.Link, error)
	list   func(ctx context.Context, limit, offset int) ([]db.Link, error)
	delete func(ctx context.Context, code string) error
	visit  func(ctx context.Context, code string) (db.Link, error)
}

func (f *fakeLinksStore) Create(ctx context.Context, url, code string) (db.Link, error) {
	return f.create(ctx, url, code)
}

func (f *fakeLinksStore) Get(ctx context.Context, code string) (db.Link, error) {
	return f.get(ctx, code)
}

func (f *fakeLinksStore) List(ctx context.Context, limit, offset int) ([]db.Link, error) {
	return f.list(ctx, limit, offset)
}

func (f *fakeLinksStore) Delete(ctx context.Context, code string) error {
	return f.delete(ctx, code)
}

func (f *fakeLinksStore) Visit(ctx context.Context, code string) (db.Link, error) {
	return f.visit(ctx, code)
}

func TestCreate_BodyTooLarge(t *testing.T) {
	a := NewLinksAPI(&fakeLinksStore{}, NewMiddleware(fakeLimiter{ok: true}))
	body := `{"url":"` + strings.Repeat("a", 1<<20) + `"}`

	req := httptest.NewRequest(http.MethodPost, "/links", strings.NewReader(body))
	rec := httptest.NewRecorder()
	a.Routes().ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusRequestEntityTooLarge; got != want {
		t.Errorf("Create.Code = %d, want %d", got, want)
	}
}

func TestCreate_BadRequest(t *testing.T) {
	tests := map[string]struct {
		body     *createBody
		wantText string
	}{
		"missing body": {
			wantText: "request body is required",
		},
		"invalid url scheme": {
			body:     &createBody{URL: "http://example.com", Code: "code"},
			wantText: "invalid url",
		},
		"invalid url host": {
			body:     &createBody{URL: "https://", Code: "code"},
			wantText: "invalid url",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			a := NewLinksAPI(&fakeLinksStore{}, NewMiddleware(fakeLimiter{ok: true}))

			var body io.Reader
			if tt.body != nil {
				b, err := json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("json.Marshal = %v", err)
				}
				body = bytes.NewReader(b)
			}

			req := httptest.NewRequest(http.MethodPost, "/links", body)
			rec := httptest.NewRecorder()
			a.Routes().ServeHTTP(rec, req)

			if got, want := rec.Code, http.StatusBadRequest; got != want {
				t.Errorf("Create.Code = %d, want %d", got, want)
			}
			if got := rec.Body.String(); !strings.Contains(got, tt.wantText) {
				t.Errorf("Create.Body = %s, want it to contain %s", got, tt.wantText)
			}
		})
	}
}

func TestCreate_StoreErrors(t *testing.T) {
	tests := map[string]struct {
		err      error
		wantCode int
	}{
		"code already exists":   {db.ErrAlreadyExists, http.StatusConflict},
		"internal server error": {errors.New("err"), http.StatusInternalServerError},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := &fakeLinksStore{}
			s.create = func(_ context.Context, _, _ string) (db.Link, error) {
				return db.Link{}, tt.err
			}
			a := NewLinksAPI(s, NewMiddleware(fakeLimiter{ok: true}))

			body := createBody{URL: "https://example.com", Code: "code"}
			b, err := json.Marshal(body)
			if err != nil {
				t.Fatalf("json.Marshal: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewReader(b))
			rec := httptest.NewRecorder()
			a.Routes().ServeHTTP(rec, req)

			if got, want := rec.Code, tt.wantCode; got != want {
				t.Errorf("Create.Code = %d, want %d", got, want)
			}
		})
	}
}

func TestCreate_Success(t *testing.T) {
	s := &fakeLinksStore{}
	var gotURL, gotCode string
	s.create = func(_ context.Context, url, code string) (db.Link, error) {
		gotURL, gotCode = url, code
		return db.Link{URL: url, Code: code}, nil
	}
	a := NewLinksAPI(s, NewMiddleware(fakeLimiter{ok: true}))

	body := createBody{URL: "https://example.com", Code: "code"}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewReader(b))
	rec := httptest.NewRecorder()
	a.Routes().ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusCreated; got != want {
		t.Errorf("Create.Code = %d, want %d", got, want)
	}

	var res createResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("json.Decode = %v", err)
	}
	if got, want := res.Link, "https://website/code"; got != want {
		t.Errorf("Create.Link = %s, want %s", got, want)
	}
	if gotURL != "https://example.com" || gotCode != "code" {
		t.Errorf("Create(url, code) = (%s, %s), want (https://example.com, code)", gotURL, gotCode)
	}
}

func TestGet_StoreErrors(t *testing.T) {
	tests := map[string]struct {
		err      error
		wantCode int
	}{
		"code not found":        {db.ErrNotFound, http.StatusNotFound},
		"internal server error": {errors.New("err"), http.StatusInternalServerError},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := &fakeLinksStore{}
			s.get = func(_ context.Context, _ string) (db.Link, error) {
				return db.Link{}, tt.err
			}
			a := NewLinksAPI(s, NewMiddleware(fakeLimiter{ok: true}))

			req := httptest.NewRequest(http.MethodGet, "/links/code", nil)
			rec := httptest.NewRecorder()
			a.Routes().ServeHTTP(rec, req)

			if got, want := rec.Code, tt.wantCode; got != want {
				t.Errorf("Get.Code = %d, want %d", got, want)
			}
		})
	}
}

func TestGet_Success(t *testing.T) {
	s := &fakeLinksStore{}
	var gotCode string
	s.get = func(_ context.Context, code string) (db.Link, error) {
		gotCode = code
		return db.Link{URL: "url"}, nil
	}
	a := NewLinksAPI(s, NewMiddleware(fakeLimiter{ok: true}))

	req := httptest.NewRequest(http.MethodGet, "/links/code", nil)
	rec := httptest.NewRecorder()
	a.Routes().ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusOK; got != want {
		t.Errorf("Get.Code = %d, want %d", got, want)
	}

	var res linkResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("json.Decode = %v", err)
	}
	if got, want := res.URL, "url"; got != want {
		t.Errorf("Get.URL = %s, want %s", got, want)
	}
	if gotCode != "code" {
		t.Errorf("Get(code) = (%s), want (code)", gotCode)
	}
}

func TestList_BadRequest(t *testing.T) {
	tests := map[string]struct {
		query        string
		wantLocation string
	}{
		"empty limit":     {"limit=0&offset=0", "query.limit"},
		"negative offset": {"limit=10&offset=-1", "query.offset"},
		"invalid limit":   {"limit=a&offset=0", "query.limit"},
		"invalid offset":  {"limit=10&offset=a", "query.offset"},
		"big limit":       {"limit=101&offset=0", "query.limit"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			a := NewLinksAPI(&fakeLinksStore{}, NewMiddleware(fakeLimiter{ok: true}))

			req := httptest.NewRequest(http.MethodGet, "/links?"+tt.query, nil)
			rec := httptest.NewRecorder()
			a.Routes().ServeHTTP(rec, req)

			if got, want := rec.Code, http.StatusUnprocessableEntity; got != want {
				t.Errorf("List.Code = %d, want %d", got, want)
			}
			if got := rec.Body.String(); !strings.Contains(got, tt.wantLocation) {
				t.Errorf("List.Body = %s, want it to contain %s", got, tt.wantLocation)
			}
		})
	}
}

func TestList_InternalServerError(t *testing.T) {
	s := &fakeLinksStore{}
	s.list = func(_ context.Context, _, _ int) ([]db.Link, error) {
		return nil, errors.New("error")
	}
	a := NewLinksAPI(s, NewMiddleware(fakeLimiter{ok: true}))

	req := httptest.NewRequest(http.MethodGet, "/links?limit=5&offset=0", nil)
	rec := httptest.NewRecorder()
	a.Routes().ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusInternalServerError; got != want {
		t.Errorf("List.Code = %d, want %d", got, want)
	}
}

func TestList_Success(t *testing.T) {
	links := make([]db.Link, 0, 5)
	for i := range 5 {
		links = append(links, db.Link{URL: fmt.Sprintf("url%d", i)})
	}

	tests := map[string]struct {
		query string
	}{
		"missing limit":          {"offset=0"},
		"missing offset":         {"limit=20"},
		"valid limit and offset": {"limit=20&offset=0"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {

			s := &fakeLinksStore{}
			var gotLimit, gotOffset int
			s.list = func(_ context.Context, limit, offset int) ([]db.Link, error) {
				gotLimit, gotOffset = limit, offset
				return links, nil
			}
			a := NewLinksAPI(s, NewMiddleware(fakeLimiter{ok: true}))

			req := httptest.NewRequest(http.MethodGet, "/links?"+tt.query, nil)
			rec := httptest.NewRecorder()
			a.Routes().ServeHTTP(rec, req)

			if got, want := rec.Code, http.StatusOK; got != want {
				t.Errorf("List.Code = %d, want %d", got, want)
			}

			var res linksResponse
			if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
				t.Fatalf("json.Decode = %v", err)
			}
			if got, want := len(res.Links), 5; got != want {
				t.Errorf("len(List) = %d, want %d", got, want)
			}
			if gotLimit != 20 || gotOffset != 0 {
				t.Errorf("List(limit, offset) = (%d, %d), want (20, 0)", gotLimit, gotOffset)
			}
		})
	}
}

func TestDelete_StoreErrors(t *testing.T) {
	tests := map[string]struct {
		err      error
		wantCode int
	}{
		"code not found":        {db.ErrNotFound, http.StatusNotFound},
		"internal server error": {errors.New("err"), http.StatusInternalServerError},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := &fakeLinksStore{}
			s.delete = func(_ context.Context, _ string) error {
				return tt.err
			}
			a := NewLinksAPI(s, NewMiddleware(fakeLimiter{ok: true}))

			req := httptest.NewRequest(http.MethodDelete, "/links/code", nil)
			rec := httptest.NewRecorder()
			a.Routes().ServeHTTP(rec, req)

			if got, want := rec.Code, tt.wantCode; got != want {
				t.Errorf("Delete.Code = %d, want %d", got, want)
			}
		})
	}
}

func TestDelete_Success(t *testing.T) {
	s := &fakeLinksStore{}
	var gotCode string
	s.delete = func(_ context.Context, code string) error {
		gotCode = code
		return nil
	}
	a := NewLinksAPI(s, NewMiddleware(fakeLimiter{ok: true}))

	req := httptest.NewRequest(http.MethodDelete, "/links/code", nil)
	rec := httptest.NewRecorder()
	a.Routes().ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusNoContent; got != want {
		t.Errorf("Delete.Code = %d, want %d", got, want)
	}
	if gotCode != "code" {
		t.Errorf("Delete(code) = (%s), want (code)", gotCode)
	}
}

func TestVisit_StoreErrors(t *testing.T) {
	tests := map[string]struct {
		err      error
		wantCode int
	}{
		"code not found":        {db.ErrNotFound, http.StatusNotFound},
		"internal server error": {errors.New("err"), http.StatusInternalServerError},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			s := &fakeLinksStore{}
			s.visit = func(_ context.Context, _ string) (db.Link, error) {
				return db.Link{}, tt.err
			}
			a := NewLinksAPI(s, NewMiddleware(fakeLimiter{ok: true}))

			req := httptest.NewRequest(http.MethodGet, "/code", nil)
			rec := httptest.NewRecorder()
			a.Routes().ServeHTTP(rec, req)

			if got, want := rec.Code, tt.wantCode; got != want {
				t.Errorf("Visit.Code = %d, want %d", got, want)
			}
		})
	}
}

func TestVisit_Success(t *testing.T) {
	s := &fakeLinksStore{}
	var gotCode string
	s.visit = func(_ context.Context, code string) (db.Link, error) {
		gotCode = code
		return db.Link{URL: "https://example.com"}, nil
	}
	a := NewLinksAPI(s, NewMiddleware(fakeLimiter{ok: true}))

	req := httptest.NewRequest(http.MethodGet, "/code", nil)
	rec := httptest.NewRecorder()
	a.Routes().ServeHTTP(rec, req)

	if got, want := rec.Code, http.StatusFound; got != want {
		t.Errorf("Visit.Code = %d, want %d", got, want)
	}
	if got, want := rec.Header().Get("Location"), "https://example.com"; got != want {
		t.Errorf("Visit.Location = %s, want %s", got, want)
	}
	if gotCode != "code" {
		t.Errorf("Visit(code) = (%s), want (code)", gotCode)
	}
}

// Package chiapi provides the implementation of the links API.
package chiapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dajimenezriv/personal/internal/db"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type linksStore interface {
	Create(ctx context.Context, url, code string) (db.Link, error)
	Get(ctx context.Context, code string) (db.Link, error)
	List(ctx context.Context, limit, offset int) ([]db.Link, error)
	Delete(ctx context.Context, code string) error
	Visit(ctx context.Context, code string) (db.Link, error)
}

type LinksAPI struct {
	s linksStore
	m *Middleware
}

func (a *LinksAPI) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(
		middleware.Heartbeat("/healthz"),
		middleware.RequestID,
		// Assumes the server is only reachable through the edge proxy, which
		// overwrites X-Real-IP with the address of the connection it accepted.
		middleware.ClientIPFromHeader("X-Real-IP"),
		traceMiddleware,
		loggerMiddleware,
		recovererMiddleware,
		// Load shedding, not rate limiting: caps in-flight requests across all
		// clients and rejects the rest straight away. 503 instead of the default
		// 429, since the server is overloaded, not the client misbehaving.
		middleware.ThrottleWithOpts(middleware.ThrottleOpts{
			Limit:        1000,
			StatusCode:   http.StatusServiceUnavailable,
			RetryAfterFn: func(bool) time.Duration { return time.Second },
		}),
		middleware.RequestSize(1<<20),
		timeoutMiddleware(10*time.Second),
	)

	r.Get("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		// Here we should ping database and more.
		w.WriteHeader(http.StatusOK)
	})

	// Serves the spec at /openapi.json and the docs UI at /docs.
	api := humachi.New(r, huma.DefaultConfig("Links API", "1.0.0"))

	huma.Register(api, huma.Operation{
		OperationID:   "create-link",
		Method:        http.MethodPost,
		Path:          "/links",
		Summary:       "Create a short link",
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusBadRequest, http.StatusConflict},
	}, a.create)
	huma.Register(api, huma.Operation{
		OperationID: "list-links",
		Method:      http.MethodGet,
		Path:        "/links",
		Summary:     "List links",
	}, a.list)
	huma.Register(api, huma.Operation{
		OperationID: "get-link",
		Method:      http.MethodGet,
		Path:        "/links/{code}",
		Summary:     "Get a link",
		Errors:      []int{http.StatusNotFound},
	}, a.get)
	huma.Register(api, huma.Operation{
		OperationID:   "delete-link",
		Method:        http.MethodDelete,
		Path:          "/links/{code}",
		Summary:       "Delete a link",
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusNotFound},
	}, a.delete)
	huma.Register(api, huma.Operation{
		OperationID:   "visit-link",
		Method:        http.MethodGet,
		Path:          "/{code}",
		Summary:       "Redirect to the link's url",
		DefaultStatus: http.StatusFound,
		Errors:        []int{http.StatusNotFound, http.StatusTooManyRequests},
		Middlewares:   huma.Middlewares{a.m.Limit("visit")},
	}, a.visit)

	return r
}

func NewLinksAPI(s linksStore, m *Middleware) *LinksAPI {
	return &LinksAPI{
		s: s,
		m: m}
}

func toLink(l db.Link) linkResponse {
	return linkResponse{URL: l.URL, Code: l.Code, Clicks: l.Clicks, CreatedAt: l.CreatedAt}
}

type linkResponse struct {
	URL       string    `json:"url"`
	Code      string    `json:"code"`
	Clicks    int       `json:"clicks"`
	CreatedAt time.Time `json:"created_at"`
}

type linksResponse struct {
	Links []linkResponse `json:"links"`
}

type codeInput struct {
	Code string `path:"code"`
}

type createInput struct {
	Body createBody
}

type createBody struct {
	URL  string `json:"url" doc:"Destination, must be an https url." example:"https://example.com"`
	Code string `json:"code,omitempty" doc:"Short code. A random one is generated if empty."`
}

type createOutput struct {
	Body createResponse
}

type createResponse struct {
	Link string `json:"link"`
}

func (a *LinksAPI) create(ctx context.Context, in *createInput) (*createOutput, error) {
	u, err := url.Parse(in.Body.URL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return nil, huma.Error400BadRequest("invalid url")
	}

	link, err := a.s.Create(ctx, in.Body.URL, in.Body.Code)
	if err != nil {
		if errors.Is(err, db.ErrAlreadyExists) {
			return nil, huma.Error409Conflict("code already taken")
		}
		return nil, toHTTPError(ctx, fmt.Errorf("links store create: %w", err))
	}

	return &createOutput{Body: createResponse{
		Link: fmt.Sprintf("https://website/%s", link.Code),
	}}, nil
}

type getOutput struct {
	Body linkResponse
}

func (a *LinksAPI) get(ctx context.Context, in *codeInput) (*getOutput, error) {
	link, err := a.s.Get(ctx, in.Code)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, huma.Error404NotFound("code not found")
		}
		return nil, toHTTPError(ctx, fmt.Errorf("links store get: %w", err))
	}

	return &getOutput{Body: toLink(link)}, nil
}

type listInput struct {
	Limit  int `query:"limit" default:"20" minimum:"1" maximum:"100"`
	Offset int `query:"offset" default:"0" minimum:"0"`
}

type listOutput struct {
	Body linksResponse
}

func (a *LinksAPI) list(ctx context.Context, in *listInput) (*listOutput, error) {
	links, err := a.s.List(ctx, in.Limit, in.Offset)
	if err != nil {
		return nil, toHTTPError(ctx, fmt.Errorf("links store list (limit %d, offset %d): %w", in.Limit, in.Offset, err))
	}

	res := linksResponse{Links: make([]linkResponse, 0, len(links))}
	for _, l := range links {
		res.Links = append(res.Links, toLink(l))
	}
	return &listOutput{Body: res}, nil
}

func (a *LinksAPI) delete(ctx context.Context, in *codeInput) (*struct{}, error) {
	if err := a.s.Delete(ctx, in.Code); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, huma.Error404NotFound("code not found")
		}
		return nil, toHTTPError(ctx, fmt.Errorf("links store delete: %w", err))
	}

	return nil, nil
}

type visitOutput struct {
	Location string `header:"Location"`
}

func (a *LinksAPI) visit(ctx context.Context, in *codeInput) (*visitOutput, error) {
	link, err := a.s.Visit(ctx, in.Code)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return nil, huma.Error404NotFound("code not found")
		}
		return nil, toHTTPError(ctx, fmt.Errorf("links store visit: %w", err))
	}

	return &visitOutput{Location: link.URL}, nil
}

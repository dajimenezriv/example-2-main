package createactor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"exercise/internal/participants/app/command/createactor"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/actor/actortest"
	"exercise/internal/testhelpers"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var actorID = uuid.MustParse("00000000-0000-0000-0000-000000000000")

func newSUT(t *testing.T, newActorID uuid.UUID) (*Handler, actor.Repository) {
	t.Helper()

	a, err := actor.New(actorID, "000000000", actor.Issuer)
	require.NoError(t, err)

	repo := actortest.NewFakeRepository()
	err = repo.Save(t.Context(), a)
	require.NoError(t, err)

	h := NewHandler(createactor.NewHandler(repo, func() uuid.UUID { return newActorID }))
	return h, repo
}

func post(t *testing.T, h *Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	api := testhelpers.NewHumaAPI(t)
	h.Register(api)
	return api.Post("/actors", strings.NewReader(body))
}

func requestBody(taxID, typ string) string {
	return fmt.Sprintf(`{"taxId": %q, "type": %q}`, taxID, typ)
}

func TestRejectedRequests(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		newActorID uuid.UUID
		wantCode   int
		wantDetail string
		wantErrors []*huma.ErrorDetail
	}{
		{
			name:       "malformed body",
			body:       ``,
			newActorID: uuid.New(),
			wantCode:   http.StatusBadRequest,
			wantDetail: "request body is required",
		},
		{
			name:       "empty body",
			body:       `{}`,
			newActorID: uuid.New(),
			wantCode:   http.StatusUnprocessableEntity,
			wantDetail: "validation failed",
			wantErrors: []*huma.ErrorDetail{
				{Message: "expected required property taxId to be present", Location: "body", Value: map[string]any{}},
				{Message: "expected required property type to be present", Location: "body", Value: map[string]any{}},
			},
		},
		{
			name:       "invalid tax ID",
			body:       requestBody("invalid", "issuer"),
			newActorID: uuid.New(),
			wantCode:   http.StatusUnprocessableEntity,
			wantDetail: "validation failed",
			wantErrors: []*huma.ErrorDetail{
				{Message: `expected string to be 9 digits`, Location: "body.taxId", Value: "invalid"},
			},
		},
		{
			name:       "invalid type",
			body:       requestBody("123456789", "invalid"),
			newActorID: uuid.New(),
			wantCode:   http.StatusUnprocessableEntity,
			wantDetail: "validation failed",
			wantErrors: []*huma.ErrorDetail{
				{Message: `expected value to be one of "issuer, investor, platform_admin, market_operator, settlement_operator, custom"`, Location: "body.type", Value: "invalid"},
			},
		},
		{
			name:       "actor already exists",
			body:       requestBody("123456789", "issuer"),
			newActorID: actorID,
			wantCode:   http.StatusConflict,
			wantDetail: "actor already exists",
		},
		{
			name:       "tax ID already exists",
			body:       requestBody("000000000", "issuer"),
			newActorID: uuid.New(),
			wantCode:   http.StatusConflict,
			wantDetail: "actor with this tax ID already exists",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, repo := newSUT(t, tt.newActorID)
			rec := post(t, h, tt.body)

			testhelpers.AssertHumaError(t, rec, tt.wantCode, tt.wantDetail, tt.wantErrors...)

			actors, err := repo.FindAll(t.Context())
			assert.NoError(t, err)
			assert.Len(t, actors, 1)
		})
	}
}

func TestCreatesActor(t *testing.T) {
	body := requestBody("999999999", "issuer")

	newActorID := uuid.New()
	h, repo := newSUT(t, newActorID)
	rec := post(t, h, body)

	want := fmt.Sprintf(`{
		"id": %q,
		"taxId": "999999999",
		"type": "issuer",
		"version": 1
	}`, newActorID)

	testhelpers.AssertHumaOutput(t, rec, http.StatusCreated, want)

	actors, err := repo.FindAll(t.Context())
	assert.NoError(t, err)
	assert.Len(t, actors, 2)
}

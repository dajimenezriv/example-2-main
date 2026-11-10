package createactor

import (
	"exercise/internal/participants/app/command/createactor"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/actor/actortest"
	"exercise/internal/testhelpers"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"github.com/danielgtaylor/huma/v2"
)

var existingActorID = uuid.New()

func newSUT(t *testing.T, newAccountID uuid.UUID) (*Handler, actor.Repository) {
	t.Helper()

	act, _ := actor.New(existingActorID, "Tax ID", actor.Issuer)

	actorRepo := actortest.NewFakeRepository()
	_ = actorRepo.Save(t.Context(), &act)

	h := NewHandler(createactor.NewHandler(
		actorRepo,
		func() uuid.UUID { return newAccountID },
	))
	return h, actorRepo
}

func post(t *testing.T, h *Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	api := testhelpers.NewHumaAPI(t)
	h.Register(api)
	return api.Post("/actors", strings.NewReader(body))
}

func TestRejectedRequests(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		newAccountID uuid.UUID
		wantCode     int
		wantDetail   string
		wantErrors   []*huma.ErrorDetail
	}{
		{
			name:         "malformed body",
			body:         ``,
			newAccountID: uuid.New(),
			wantCode:     http.StatusBadRequest,
			wantDetail:   "request body is required",
		},
		{
			name:         "empty body",
			body:         `{}`,
			newAccountID: uuid.New(),
			wantCode:     http.StatusUnprocessableEntity,
			wantDetail:   "validation failed",
			wantErrors: []*huma.ErrorDetail{
				{Message: "expected required property tax_id to be present", Location: "body", Value: map[string]any{}},
				{Message: "expected required property type to be present", Location: "body", Value: map[string]any{}},
			},
		},
		{
			name:         "invalid type",
			body:         `{"tax_id": "Tax ID", "type": "invalid"}`,
			newAccountID: uuid.New(),
			wantCode:     http.StatusUnprocessableEntity,
			wantDetail:   "validation failed",
			wantErrors: []*huma.ErrorDetail{
				{Message: `expected value to be one of "issuer, investor, platform_admin, market_operator, settlement_operator, custom"`, Location: "body.type", Value: "invalid"},
			},
		},
		{
			name:         "actor already exists",
			body:         `{"tax_id": "Tax ID", "type": "issuer"}`,
			newAccountID: existingActorID,
			wantCode:     http.StatusConflict,
			wantDetail:   "actor already exists",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, actorRepo := newSUT(t, tt.newAccountID)
			rec := post(t, h, tt.body)

			testhelpers.AssertHumaError(t, rec, tt.wantCode, tt.wantDetail, tt.wantErrors...)

			actors, _ := actorRepo.FindAll(t.Context())
			if got := len(actors); got != 1 {
				t.Errorf("len(actors) = %d, want 1", got)
			}
		})
	}
}

func TestCreatesActor(t *testing.T) {
	body := `{"tax_id": "Tax ID", "type": "issuer"}`

	newAccountID := uuid.New()
	h, actorRepo := newSUT(t, newAccountID)
	rec := post(t, h, body)

	want := fmt.Sprintf(`{
		"id": %q,
		"tax_id": "Tax ID",
		"type": "issuer"
	}`, newAccountID)

	testhelpers.AssertHumaOutput(t, rec, http.StatusCreated, want)

	actors, _ := actorRepo.FindAll(t.Context())
	if got := len(actors); got != 2 {
		t.Errorf("len(actors) = %d, want 2", got)
	}
}

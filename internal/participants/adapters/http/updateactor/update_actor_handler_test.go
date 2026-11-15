package updateactor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"exercise/internal/participants/app/command/updateactor"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/actor/actortest"
	"exercise/internal/testhelpers"
)

var actorID = uuid.MustParse("00000000-0000-0000-0000-000000000000")

func saveActor(t *testing.T, repo actor.Repository, id uuid.UUID, taxID string) *actor.Actor {
	t.Helper()
	a, err := actor.New(id, taxID, actor.Issuer)
	require.NoError(t, err)
	err = repo.Save(t.Context(), a)
	require.NoError(t, err)
	return a
}

func newSUT(t *testing.T) (*actor.Actor, *Handler, actor.Repository) {
	t.Helper()

	repo := actortest.NewFakeRepository()
	a := saveActor(t, repo, actorID, "000000000")
	saveActor(t, repo, uuid.New(), "111111111")

	h := NewHandler(updateactor.NewHandler(repo))
	return a, h, repo
}

func patch(t *testing.T, h *Handler, id string, body string) *httptest.ResponseRecorder {
	t.Helper()
	api := testhelpers.NewHumaAPI(t)
	h.Register(api)
	return api.Patch("/actors/"+id, strings.NewReader(body))
}

func requestBody(taxID, typ string, version int64) string {
	return fmt.Sprintf(`{"taxId": %q, "type": %q, "version": %d}`, taxID, typ, version)
}

func TestRejectedRequests(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		body       string
		wantCode   int
		wantDetail string
		wantErrors []*huma.ErrorDetail
	}{
		{
			name:       "malformed body",
			body:       ``,
			id:         actorID.String(),
			wantCode:   http.StatusBadRequest,
			wantDetail: "request body is required",
		},
		{
			name:       "empty body",
			id:         actorID.String(),
			body:       `{}`,
			wantCode:   http.StatusUnprocessableEntity,
			wantDetail: "validation failed",
			wantErrors: []*huma.ErrorDetail{
				{Message: "expected required property version to be present", Location: "body", Value: map[string]any{}},
			},
		},
		{
			name:       "invalid tax ID",
			id:         actorID.String(),
			body:       requestBody("invalid", "issuer", 1),
			wantCode:   http.StatusUnprocessableEntity,
			wantDetail: "validation failed",
			wantErrors: []*huma.ErrorDetail{
				{Message: `expected string to be 9 digits`, Location: "body.taxId", Value: "invalid"},
			},
		},
		{
			name:       "invalid type",
			id:         actorID.String(),
			body:       requestBody("000000000", "invalid", 1),
			wantCode:   http.StatusUnprocessableEntity,
			wantDetail: "validation failed",
			wantErrors: []*huma.ErrorDetail{
				{Message: `expected value to be one of "issuer, investor, platform_admin, market_operator, settlement_operator, custom"`, Location: "body.type", Value: "invalid"},
			},
		},
		{
			name:       "non positive version",
			id:         actorID.String(),
			body:       requestBody("000000000", "issuer", 0),
			wantCode:   http.StatusUnprocessableEntity,
			wantDetail: "validation failed",
			wantErrors: []*huma.ErrorDetail{
				{Message: `expected number >= 1`, Location: "body.version", Value: float64(0)},
			},
		},
		{
			// Huma accepts an uppercase "URN:UUID:" prefix, but uuid.Parse does not.
			name:       "invalid actor ID",
			id:         "URN:UUID:" + actorID.String(),
			body:       requestBody("000000000", "issuer", 1),
			wantCode:   http.StatusBadRequest,
			wantDetail: "invalid actorId",
		},
		{
			name:       "actor does not exist",
			id:         uuid.New().String(),
			body:       requestBody("000000000", "issuer", 1),
			wantCode:   http.StatusNotFound,
			wantDetail: "actor not found",
		},
		{
			name:       "tax ID already exists",
			id:         actorID.String(),
			body:       requestBody("111111111", "issuer", 1),
			wantCode:   http.StatusConflict,
			wantDetail: "actor with this tax ID already exists",
		},
		{
			name:       "version conflict",
			id:         actorID.String(),
			body:       requestBody("000000000", "issuer", 2),
			wantCode:   http.StatusConflict,
			wantDetail: "actor version conflict",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, h, repo := newSUT(t)
			rec := patch(t, h, tt.id, tt.body)

			testhelpers.AssertHumaError(t, rec, tt.wantCode, tt.wantDetail, tt.wantErrors...)

			got, err := repo.FindByID(t.Context(), a.ID())
			assert.NoError(t, err)
			assert.Equal(t, a, got)
		})
	}
}

func TestUpdatesActor(t *testing.T) {
	tests := []struct {
		name      string
		body      string
		wantTaxID string
		wantType  actor.Type
	}{
		{
			name:      "tax ID and type",
			body:      requestBody("999999999", "investor", 1),
			wantTaxID: "999999999",
			wantType:  actor.Investor,
		},
		{
			name:      "tax ID only",
			body:      `{"taxId": "999999999", "version": 1}`,
			wantTaxID: "999999999",
			wantType:  actor.Issuer,
		},
		{
			name:      "type only",
			body:      `{"type": "investor", "version": 1}`,
			wantTaxID: "000000000",
			wantType:  actor.Investor,
		},
		{
			name:      "neither",
			body:      `{"version": 1}`,
			wantTaxID: "000000000",
			wantType:  actor.Issuer,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, h, repo := newSUT(t)
			rec := patch(t, h, actorID.String(), tt.body)

			want := fmt.Sprintf(`{
				"id": %q,
				"taxId": %q,
				"type": %q,
				"version": 2
			}`, actorID, tt.wantTaxID, tt.wantType)

			testhelpers.AssertHumaOutput(t, rec, http.StatusOK, want)

			got, err := repo.FindByID(t.Context(), actorID)
			assert.NoError(t, err)
			wantActor := actor.Reconstitute(actorID, tt.wantTaxID, tt.wantType, 2)
			assert.Equal(t, wantActor, got)
		})
	}
}

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

var actorID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

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

func patch(t *testing.T, h *Handler, id, ifMatch, body string) *httptest.ResponseRecorder {
	t.Helper()
	api := testhelpers.NewHumaAPI(t)
	h.Register(api)
	args := []any{strings.NewReader(body)}
	if ifMatch != "" {
		args = append(args, "If-Match: "+ifMatch)
	}
	return api.Patch("/actors/"+id, args...)
}

func requestBody(taxID, typ string) string {
	return fmt.Sprintf(`{"taxId": %q, "type": %q}`, taxID, typ)
}

func TestRejectedRequests(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		ifMatch    string
		body       string
		wantCode   int
		wantDetail string
		wantErrors []*huma.ErrorDetail
	}{
		{
			name:       "malformed body",
			id:         actorID.String(),
			ifMatch:    `"1"`,
			body:       ``,
			wantCode:   http.StatusBadRequest,
			wantDetail: "request body is required",
		},
		{
			name:       "invalid tax ID",
			id:         actorID.String(),
			ifMatch:    `"1"`,
			body:       requestBody("invalid", "issuer"),
			wantCode:   http.StatusUnprocessableEntity,
			wantDetail: "validation failed",
			wantErrors: []*huma.ErrorDetail{
				{Message: `expected string to be 9 digits`, Location: "body.taxId", Value: "invalid"},
			},
		},
		{
			name:       "invalid type",
			id:         actorID.String(),
			ifMatch:    `"1"`,
			body:       requestBody("000000000", "invalid"),
			wantCode:   http.StatusUnprocessableEntity,
			wantDetail: "validation failed",
			wantErrors: []*huma.ErrorDetail{
				{Message: `expected value to be one of "issuer, investor, platform_admin, market_operator, settlement_operator, custom"`, Location: "body.type", Value: "invalid"},
			},
		},
		{
			name:       "missing If-Match",
			id:         actorID.String(),
			body:       requestBody("000000000", "issuer"),
			wantCode:   http.StatusPreconditionRequired,
			wantDetail: "If-Match header is required",
		},
		{
			name:       "malformed If-Match",
			id:         actorID.String(),
			ifMatch:    `W/"1"`,
			body:       requestBody("000000000", "issuer"),
			wantCode:   http.StatusBadRequest,
			wantDetail: `If-Match must be a single strong ETag, e.g. "1"`,
		},
		{
			// Huma accepts an uppercase "URN:UUID:" prefix, but uuid.Parse does not.
			name:       "invalid actor ID",
			id:         "URN:UUID:" + actorID.String(),
			ifMatch:    `"1"`,
			body:       requestBody("000000000", "issuer"),
			wantCode:   http.StatusBadRequest,
			wantDetail: "invalid id",
		},
		{
			name:       "actor does not exist",
			id:         uuid.New().String(),
			ifMatch:    `"1"`,
			body:       requestBody("000000000", "issuer"),
			wantCode:   http.StatusNotFound,
			wantDetail: "actor not found",
		},
		{
			name:       "tax ID already exists",
			id:         actorID.String(),
			ifMatch:    `"1"`,
			body:       requestBody("111111111", "issuer"),
			wantCode:   http.StatusConflict,
			wantDetail: "actor with this tax ID already exists",
		},
		{
			name:       "version conflict",
			id:         actorID.String(),
			ifMatch:    `"2"`,
			body:       requestBody("000000000", "issuer"),
			wantCode:   http.StatusPreconditionFailed,
			wantDetail: "actor version conflict",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, h, repo := newSUT(t)
			rec := patch(t, h, tt.id, tt.ifMatch, tt.body)

			testhelpers.AssertHumaError(t, rec, tt.wantCode, tt.wantDetail, tt.wantErrors...)

			got, err := repo.FindByID(t.Context(), a.ID())
			require.NoError(t, err)
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
			body:      requestBody("999999999", "investor"),
			wantTaxID: "999999999",
			wantType:  actor.Investor,
		},
		{
			name:      "tax ID only",
			body:      `{"taxId": "999999999"}`,
			wantTaxID: "999999999",
			wantType:  actor.Issuer,
		},
		{
			name:      "type only",
			body:      `{"type": "investor"}`,
			wantTaxID: "000000000",
			wantType:  actor.Investor,
		},
		{
			name:      "neither",
			body:      `{}`,
			wantTaxID: "000000000",
			wantType:  actor.Issuer,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, h, repo := newSUT(t)
			rec := patch(t, h, actorID.String(), `"1"`, tt.body)

			want := fmt.Sprintf(`{
				"id": %q,
				"taxId": %q,
				"type": %q,
				"version": 2
			}`, actorID, tt.wantTaxID, tt.wantType)

			testhelpers.AssertHumaOutput(t, rec, http.StatusOK, want)
			assert.Equal(t, `"2"`, rec.Header().Get("ETag"))

			got, err := repo.FindByID(t.Context(), actorID)
			require.NoError(t, err)
			wantActor := actor.Reconstitute(actorID, tt.wantTaxID, tt.wantType, 2)
			assert.Equal(t, wantActor, got)
		})
	}
}

package deleteactor

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"exercise/internal/participants/app/command/deleteactor"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/actor/actortest"
	"exercise/internal/testhelpers"
)

var actorID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

func newSUT(t *testing.T) (*Handler, actor.Repository) {
	t.Helper()

	repo := actortest.NewFakeRepository()
	a, err := actor.New(actorID, "000000000", actor.Issuer)
	require.NoError(t, err)
	require.NoError(t, repo.Save(t.Context(), a))

	return NewHandler(deleteactor.NewHandler(repo)), repo
}

func del(t *testing.T, h *Handler, id, ifMatch string) *httptest.ResponseRecorder {
	t.Helper()
	api := testhelpers.NewHumaAPI(t)
	h.Register(api)
	if ifMatch == "" {
		return api.Delete("/actors/" + id)
	}
	return api.Delete("/actors/"+id, "If-Match: "+ifMatch)
}

func TestRejectedRequests(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		ifMatch    string
		wantCode   int
		wantDetail string
	}{
		{
			name:       "missing If-Match",
			id:         actorID.String(),
			wantCode:   http.StatusPreconditionRequired,
			wantDetail: "If-Match header is required",
		},
		{
			name:       "malformed If-Match",
			id:         actorID.String(),
			ifMatch:    `1`,
			wantCode:   http.StatusBadRequest,
			wantDetail: `If-Match must be a single strong ETag, e.g. "1"`,
		},
		{
			name:       "wildcard If-Match",
			id:         actorID.String(),
			ifMatch:    `*`,
			wantCode:   http.StatusBadRequest,
			wantDetail: `If-Match must be a single strong ETag, e.g. "1"`,
		},
		{
			// Huma accepts an uppercase "URN:UUID:" prefix, but uuid.Parse does not.
			name:       "invalid actor ID",
			id:         "URN:UUID:" + actorID.String(),
			ifMatch:    `"1"`,
			wantCode:   http.StatusBadRequest,
			wantDetail: "invalid id",
		},
		{
			name:       "actor does not exist",
			id:         uuid.New().String(),
			ifMatch:    `"1"`,
			wantCode:   http.StatusNotFound,
			wantDetail: "actor not found",
		},
		{
			name:       "version conflict",
			id:         actorID.String(),
			ifMatch:    `"2"`,
			wantCode:   http.StatusPreconditionFailed,
			wantDetail: "actor version conflict",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, repo := newSUT(t)
			rec := del(t, h, tt.id, tt.ifMatch)

			testhelpers.AssertHumaError(t, rec, tt.wantCode, tt.wantDetail)

			_, err := repo.FindByID(t.Context(), actorID)
			require.NoError(t, err, "actor must survive a rejected delete")
		})
	}
}

func TestDeletesActor(t *testing.T) {
	h, repo := newSUT(t)

	rec := del(t, h, actorID.String(), `"1"`)

	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Body.String())
	_, err := repo.FindByID(t.Context(), actorID)
	require.ErrorIs(t, err, actor.ErrNotFound)
}

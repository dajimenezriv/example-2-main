package getactors

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"github.com/stretchr/testify/require"

	"exercise/internal/participants/app/query/getactors"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/actor/actortest"
	"exercise/internal/testhelpers"
)

var (
	actorID1   = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	actorID2   = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	actorID3   = uuid.MustParse("00000000-0000-0000-0000-000000000003")
	accountID1 = uuid.MustParse("00000000-0000-0000-0000-000000000004")
	accountID2 = uuid.MustParse("00000000-0000-0000-0000-000000000005")
	accountID3 = uuid.MustParse("00000000-0000-0000-0000-000000000006")
)

func newSUT(t *testing.T) (actor.Repository, account.Repository) {
	t.Helper()
	actorRepo := actortest.NewFakeRepository()
	saveActor(t, actorRepo, actorID1, "000000000")
	saveActor(t, actorRepo, actorID2, "000000001")
	saveActor(t, actorRepo, actorID3, "000000002")
	accountRepo := accounttest.NewFakeRepository()
	saveAccount(t, accountRepo, accountID1, actorID1)
	saveAccount(t, accountRepo, accountID2, actorID1)
	saveAccount(t, accountRepo, accountID3, actorID2)
	return actorRepo, accountRepo
}

func saveActor(t *testing.T, repo actor.Repository, id uuid.UUID, taxID string) {
	t.Helper()
	a, err := actor.New(id, taxID, actor.Issuer)
	require.NoError(t, err)
	err = repo.Save(context.Background(), a)
	require.NoError(t, err)
}

func saveAccount(t *testing.T, repo account.Repository, id, actorID uuid.UUID) {
	t.Helper()
	a, err := account.New(id, actorID, "dltID", "privKey", "pubKey")
	require.NoError(t, err)
	err = repo.Save(context.Background(), a)
	require.NoError(t, err)
}

func get(ctx context.Context, t *testing.T, actorRepo actor.Repository, accountRepo account.Repository, url string) *httptest.ResponseRecorder {
	t.Helper()
	h := NewHandler(getactors.NewQuery(actorRepo, accountRepo))
	api := testhelpers.NewHumaAPI(t)
	h.Register(api)
	return api.GetCtx(ctx, url)
}

var (
	actor1JSON = fmt.Sprintf(`{"id": %q, "taxId": "000000000", "type": "issuer", "version": 1, "accounts": [
		{"id": %q, "dltAccountId": "dltID", "publicKey": "pubKey"},
		{"id": %q, "dltAccountId": "dltID", "publicKey": "pubKey"}
	]}`, actorID1, accountID1, accountID2)
	actor2JSON = fmt.Sprintf(`{"id": %q, "taxId": "000000001", "type": "issuer", "version": 1, "accounts": [
		{"id": %q, "dltAccountId": "dltID", "publicKey": "pubKey"}
	]}`, actorID2, accountID3)
	actor3JSON = fmt.Sprintf(`{"id": %q, "taxId": "000000002", "type": "issuer", "version": 1, "accounts": []}`, actorID3)
)

func page(hasNext, hasPrevious bool, actors ...string) string {
	return fmt.Sprintf(
		`{"actors": [%s], "totalCount": 3, "hasNext": %t, "hasPrevious": %t}`,
		strings.Join(actors, ", "), hasNext, hasPrevious,
	)
}

func TestGetActors_RejectedRequests(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	tests := []struct {
		name       string
		ctx        context.Context
		url        string
		wantCode   int
		wantDetail string
	}{
		{
			name:       "context canceled",
			ctx:        ctx,
			url:        "/actors",
			wantCode:   http.StatusInternalServerError,
			wantDetail: "internal server error",
		},
		{
			name:       "zero limit",
			ctx:        t.Context(),
			url:        "/actors?limit=0",
			wantCode:   http.StatusBadRequest,
			wantDetail: "invalid pagination",
		},
		{
			name:       "limit above max",
			ctx:        t.Context(),
			url:        "/actors?limit=21",
			wantCode:   http.StatusBadRequest,
			wantDetail: "invalid pagination",
		},
		{
			name:       "negative offset",
			ctx:        t.Context(),
			url:        "/actors?offset=-1",
			wantCode:   http.StatusBadRequest,
			wantDetail: "invalid pagination",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actorRepo, accountRepo := newSUT(t)
			rec := get(tt.ctx, t, actorRepo, accountRepo, tt.url)

			testhelpers.AssertHumaError(t, rec, tt.wantCode, tt.wantDetail)
		})
	}
}

func TestGetActors_Success(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "no limit or offset",
			url:  "/actors",
			want: page(false, false, actor1JSON, actor2JSON, actor3JSON),
		},
		{
			name: "max limit",
			url:  "/actors?limit=20",
			want: page(false, false, actor1JSON, actor2JSON, actor3JSON),
		},
		{
			name: "set limit",
			url:  "/actors?limit=2",
			want: page(true, false, actor1JSON, actor2JSON),
		},
		{
			name: "set limit and offset",
			url:  "/actors?limit=1&offset=1",
			want: page(true, true, actor2JSON),
		},
		{
			name: "last page",
			url:  "/actors?limit=2&offset=2",
			want: page(false, true, actor3JSON),
		},
		{
			name: "offset past end",
			url:  "/actors?offset=10",
			want: page(false, true),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actorRepo, accountRepo := newSUT(t)
			rec := get(t.Context(), t, actorRepo, accountRepo, tt.url)

			testhelpers.AssertHumaOutput(t, rec, http.StatusOK, tt.want)
		})
	}
}

package getactors

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func post(ctx context.Context, t *testing.T, actorRepo actor.Repository, accountRepo account.Repository) *httptest.ResponseRecorder {
	t.Helper()
	h := NewHandler(getactors.NewQuery(actorRepo, accountRepo))
	api := testhelpers.NewHumaAPI(t)
	h.Register(api)
	return api.GetCtx(ctx, "/actors")
}

func TestContextCanceledReturnsError(t *testing.T) {
	actorRepo, accountRepo := newSUT(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	rec := post(ctx, t, actorRepo, accountRepo)

	testhelpers.AssertHumaError(t, rec, http.StatusInternalServerError, "find actors: context canceled")
}

func TestReturnActorsAndAccounts(t *testing.T) {
	actorRepo, accountRepo := newSUT(t)
	rec := post(t.Context(), t, actorRepo, accountRepo)

	want := fmt.Sprintf(`{"actors": [
		{"id": %q, "taxId": "000000000", "type": "issuer", "version": 1, "accounts": [
			{"id": %q, "dltAccountId": "dltID", "publicKey": "pubKey"},
			{"id": %q, "dltAccountId": "dltID", "publicKey": "pubKey"}
		]},
		{"id": %q, "taxId": "000000001", "type": "issuer", "version": 1, "accounts": [
			{"id": %q, "dltAccountId": "dltID", "publicKey": "pubKey"}
		]},
		{"id": %q, "taxId": "000000002", "type": "issuer", "version": 1, "accounts": []}
	]}`, actorID1, accountID1, accountID2, actorID2, accountID3, actorID3)

	testhelpers.AssertHumaOutput(t, rec, http.StatusOK, want)
}

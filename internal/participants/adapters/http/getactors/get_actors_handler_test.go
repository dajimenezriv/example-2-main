package getactors

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"exercise/internal/participants/app/query/getactors"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/actor/actortest"
	"exercise/internal/testhelpers"
)

var (
	actorID1   = uuid.New()
	actorID2   = uuid.New()
	actorID3   = uuid.New()
	accountID1 = uuid.New()
	accountID2 = uuid.New()
	accountID3 = uuid.New()
)

func newSUT(t *testing.T) (actor.Repository, account.Repository) {
	t.Helper()
	actorRepo := actortest.NewFakeRepository()
	saveActor(t, actorRepo, actorID1)
	saveActor(t, actorRepo, actorID2)
	saveActor(t, actorRepo, actorID3)
	accountRepo := accounttest.NewFakeRepository()
	saveAccount(t, accountRepo, accountID1, actorID1)
	saveAccount(t, accountRepo, accountID2, actorID1)
	saveAccount(t, accountRepo, accountID3, actorID2)
	return actorRepo, accountRepo
}

func saveActor(t *testing.T, repo actor.Repository, uuid uuid.UUID) *actor.Actor {
	t.Helper()
	a, _ := actor.New(uuid, "Tax ID", actor.Issuer)
	_ = repo.Save(context.Background(), &a)
	return &a
}

func saveAccount(t *testing.T, repo account.Repository, uuid, actorID uuid.UUID) *account.Account {
	t.Helper()
	a, _ := account.New(uuid, actorID, "DLT account ID", "Private key", "Public key")
	_ = repo.Save(context.Background(), &a)
	return &a
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
		{"id": %q, "taxId": "Tax ID", "type": "issuer", "accounts": [
			{"id": %q, "dltAccountId": "DLT account ID", "publicKey": "Public key"},
			{"id": %q, "dltAccountId": "DLT account ID", "publicKey": "Public key"}
		]},
		{"id": %q, "taxId": "Tax ID", "type": "issuer", "accounts": [
			{"id": %q, "dltAccountId": "DLT account ID", "publicKey": "Public key"}
		]},
		{"id": %q, "taxId": "Tax ID", "type": "issuer", "accounts": []}
	]}`, actorID1, accountID1, accountID2, actorID2, accountID3, actorID3)

	testhelpers.AssertHumaOutput(t, rec, http.StatusOK, want)
}

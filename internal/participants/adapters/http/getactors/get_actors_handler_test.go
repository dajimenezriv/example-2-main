package getactors

import (
	"context"
	"exercise/internal/participants/app/query/getactors"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/actor/actortest"
	"exercise/internal/tests"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"
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
	a, err := actor.New(uuid, "Tax ID", actor.Issuer)
	if err != nil {
		t.Fatalf("new actor = %v", err)
	}
	if err := repo.Save(context.Background(), &a); err != nil {
		t.Fatalf("actor repo save = %v", err)
	}
	return &a
}

func saveAccount(t *testing.T, repo account.Repository, uuid, actorID uuid.UUID) *account.Account {
	t.Helper()
	a := account.New(uuid, actorID, "DLT account ID", "Private key", "Public key")
	if err := repo.Save(context.Background(), &a); err != nil {
		t.Fatalf("actor repo save = %v", err)
	}
	return &a
}

func post(t *testing.T, ctx context.Context, actorRepo actor.Repository, accountRepo account.Repository) *httptest.ResponseRecorder {
	t.Helper()
	h := NewHandler(getactors.NewQuery(actorRepo, accountRepo))
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, "GET", "/actors", nil)
	h.ServeHTTP(rec, req)
	return rec
}

func TestGetActorsQueryFails(t *testing.T) {
	actorRepo, accountRepo := newSUT(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	rec := post(t, ctx, actorRepo, accountRepo)

	if got, want := rec.Code, http.StatusInternalServerError; got != want {
		t.Errorf("Code = %d, want %d", got, want)
	}
	want := ""
	if got := rec.Body.String(); got != want {
		t.Errorf("Body = %q, want %q", got, want)
	}
}

func TestGetActorsReturnActors(t *testing.T) {
	actorRepo, accountRepo := newSUT(t)
	rec := post(t, t.Context(), actorRepo, accountRepo)

	if got, want := rec.Code, http.StatusOK; got != want {
		t.Errorf("Code = %d, want %d", got, want)
	}
	if got, want := rec.Header().Get("Content-Type"), "application/json"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}

	want := fmt.Sprintf(`{"actors": [
	{"id": %q, "tax_id": "Tax ID", "type": "issuer", "accounts": [
		{"id": %q, "dlt_account_id": "DLT account ID", "public_key": "Public key"},
		{"id": %q, "dlt_account_id": "DLT account ID", "public_key": "Public key"}
	]},
	{"id": %q, "tax_id": "Tax ID", "type": "issuer", "accounts": [
		{"id": %q, "dlt_account_id": "DLT account ID", "public_key": "Public key"}
	]},
	{"id": %q, "tax_id": "Tax ID", "type": "issuer", "accounts": []}
]}`, actorID1, accountID1, accountID2, actorID2, accountID3, actorID3)
	tests.AssertJSONEqual(t, rec.Body.String(), want)
}

package getaccounts

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"exercise/internal/participants/app/query/getaccounts"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
	"exercise/internal/testhelpers"
)

var existingAccountID = uuid.New()

func newSUT(t *testing.T) account.Repository {
	t.Helper()
	repo := accounttest.NewFakeRepository()
	a, _ := account.New(existingAccountID, uuid.New(), "DLT account ID", "Private key", "Public key")
	_ = repo.Save(t.Context(), &a)
	return repo
}

func post(ctx context.Context, t *testing.T, repo account.Repository) *httptest.ResponseRecorder {
	t.Helper()
	h := NewHandler(&getaccounts.Query{Repo: repo})
	api := testhelpers.NewHumaAPI(t)
	h.Register(api)
	return api.GetCtx(ctx, "/accounts")
}

func TestContextCanceledReturnsError(t *testing.T) {
	repo := newSUT(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	rec := post(ctx, t, repo)

	if got, want := rec.Code, http.StatusInternalServerError; got != want {
		t.Errorf("Code = %d, want %d", got, want)
	}
}

func TestReturnsAllAccounts(t *testing.T) {
	repo := newSUT(t)
	rec := post(t.Context(), t, repo)

	want := fmt.Sprintf(`{"accounts": [
		{"id": %q, "dlt_account_id": "DLT account ID", "public_key": "Public key"}
	]}`, existingAccountID.String())

	testhelpers.AssertHumaOutput(t, rec, http.StatusOK, want)
}

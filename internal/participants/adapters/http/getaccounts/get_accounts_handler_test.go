package getaccounts

import (
	"context"
	"exercise/internal/participants/app/query/getaccounts"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"
)

var existingAccountID = uuid.New()

func newSUT(t *testing.T) account.Repository {
	t.Helper()
	repo := accounttest.NewFakeRepository()
	a := account.New(existingAccountID, uuid.New(), "DLT account ID", "Private key", "Public key")
	if err := repo.Save(t.Context(), &a); err != nil {
		t.Fatalf("account repo save = %v", err)
	}
	return repo
}

func post(t *testing.T, ctx context.Context, repo account.Repository) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequestWithContext(ctx, "GET", "/accounts", nil)
	h := NewHandler(&getaccounts.Query{Repo: repo})
	h.ServeHTTP(rec, req)
	return rec
}

func TestGetAccountsQueryFails(t *testing.T) {
	repo := newSUT(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	rec := post(t, ctx, repo)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Code = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestGetAccountsReturnsAllAccounts(t *testing.T) {
	repo := newSUT(t)
	rec := post(t, t.Context(), repo)

	if rec.Code != http.StatusOK {
		t.Errorf("Code = %d, want %d", rec.Code, http.StatusOK)
	}
	if got, want := rec.Header().Get("Content-Type"), "application/json"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}

	want := fmt.Sprintf(
		`{"accounts":[{"id":%q,"dlt_account_id":"DLT account ID","public_key":"Public key"}]}`+"\n",
		existingAccountID.String(),
	)
	if got := rec.Body.String(); got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}

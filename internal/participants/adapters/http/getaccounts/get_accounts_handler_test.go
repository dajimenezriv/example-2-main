package getaccounts

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"exercise/internal/participants/app/query/getaccounts"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
	"exercise/internal/testhelpers"
)

var (
	accountID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	actorID   = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

func newSUT(t *testing.T) account.Repository {
	t.Helper()
	repo := accounttest.NewFakeRepository()
	a, err := account.New(accountID, actorID, "dltID", "privKey", "pubKey")
	require.NoError(t, err)
	err = repo.Save(t.Context(), a)
	require.NoError(t, err)
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
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestReturnsAllAccounts(t *testing.T) {
	repo := newSUT(t)
	rec := post(t.Context(), t, repo)

	want := fmt.Sprintf(`{"accounts": [
		{"id": %q, "dltAccountId": "dltID", "publicKey": "pubKey", "version": 1}
	]}`, accountID.String())

	testhelpers.AssertHumaOutput(t, rec, http.StatusOK, want)
}

package getaccounts

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"uuid"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"exercise/internal/participants/app/query/getaccounts"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
	"exercise/internal/testhelpers"
)

var (
	accountID1 = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	accountID2 = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	accountID3 = uuid.MustParse("00000000-0000-0000-0000-000000000003")
	actorID    = uuid.MustParse("00000000-0000-0000-0000-000000000004")
)

func newSUT(t *testing.T) account.Repository {
	t.Helper()
	repo := accounttest.NewFakeRepository()
	saveAccount(t, repo, accountID1)
	saveAccount(t, repo, accountID2)
	saveAccount(t, repo, accountID3)
	return repo
}

func saveAccount(t *testing.T, repo account.Repository, accountID uuid.UUID) {
	t.Helper()
	a, err := account.New(accountID, actorID, "dltID", "privKey", "pubKey")
	require.NoError(t, err)
	err = repo.Save(t.Context(), a)
	require.NoError(t, err)
}

func get(ctx context.Context, t *testing.T, repo account.Repository, url string) *httptest.ResponseRecorder {
	t.Helper()
	h := NewHandler(&getaccounts.Query{Repo: repo})
	api := testhelpers.NewHumaAPI(t)
	h.Register(api)
	return api.GetCtx(ctx, url)
}

func row(accountID uuid.UUID) string {
	return fmt.Sprintf(`{"id": %q, "dltAccountId": "dltID", "publicKey": "pubKey", "version": 1}`, accountID)
}

func TestContextCanceledReturnsError(t *testing.T) {
	repo := newSUT(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	rec := get(ctx, t, repo, "/accounts")
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetAccounts_RejectedRequests(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantCode   int
		wantDetail string
		wantErrors []*huma.ErrorDetail
	}{
		{
			name:       "zero limit",
			url:        "/accounts?limit=0",
			wantCode:   http.StatusUnprocessableEntity,
			wantDetail: "validation failed",
			wantErrors: []*huma.ErrorDetail{
				{Message: "expected number >= 1", Location: "query.limit", Value: float64(0)},
			},
		},
		{
			name:       "big limit",
			url:        "/accounts?limit=101",
			wantCode:   http.StatusUnprocessableEntity,
			wantDetail: "validation failed",
			wantErrors: []*huma.ErrorDetail{
				{Message: "expected number <= 100", Location: "query.limit", Value: float64(101)},
			},
		},
		{
			name:       "negative offset",
			url:        "/accounts?offset=-1",
			wantCode:   http.StatusUnprocessableEntity,
			wantDetail: "validation failed",
			wantErrors: []*huma.ErrorDetail{
				{Message: "expected number >= 0", Location: "query.offset", Value: float64(-1)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newSUT(t)
			rec := get(t.Context(), t, repo, tt.url)

			testhelpers.AssertHumaError(t, rec, tt.wantCode, tt.wantDetail, tt.wantErrors...)
		})
	}
}

func TestGetAccounts_Success(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "no limit or offset",
			url:  "/accounts",
			want: fmt.Sprintf(`{"accounts": [%s, %s, %s]}`, row(accountID1), row(accountID2), row(accountID3)),
		},
		{
			name: "set limit",
			url:  "/accounts?limit=2",
			want: fmt.Sprintf(`{"accounts": [%s, %s]}`, row(accountID1), row(accountID2)),
		},
		{
			name: "limit bigger than number of accounts",
			url:  "/accounts?limit=10",
			want: fmt.Sprintf(`{"accounts": [%s, %s, %s]}`, row(accountID1), row(accountID2), row(accountID3)),
		},
		{
			name: "set offset",
			url:  "/accounts?offset=1",
			want: fmt.Sprintf(`{"accounts": [%s, %s]}`, row(accountID2), row(accountID3)),
		},
		{
			name: "offset bigger than number of accounts",
			url:  "/accounts?offset=4",
			want: `{"accounts": []}`,
		},
		{
			name: "set limit and offset",
			url:  "/accounts?limit=1&&offset=1",
			want: fmt.Sprintf(`{"accounts": [%s]}`, row(accountID2)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newSUT(t)
			rec := get(t.Context(), t, repo, tt.url)

			testhelpers.AssertHumaOutput(t, rec, http.StatusOK, tt.want)
		})
	}
}

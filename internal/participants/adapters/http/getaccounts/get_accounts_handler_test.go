package getaccounts

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

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

func page(hasNext, hasPrevious bool, accountIDs ...uuid.UUID) string {
	rows := make([]string, len(accountIDs))
	for i, id := range accountIDs {
		rows[i] = fmt.Sprintf(`{"id": %q, "dltAccountId": "dltID", "publicKey": "pubKey", "version": 1}`, id)
	}
	return fmt.Sprintf(
		`{"accounts": [%s], "totalCount": 3, "hasNext": %t, "hasPrevious": %t}`,
		strings.Join(rows, ", "), hasNext, hasPrevious,
	)
}

func TestGetAccounts_RejectedRequests(t *testing.T) {
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
			url:        "/accounts",
			wantCode:   http.StatusInternalServerError,
			wantDetail: "internal server error",
		},
		{
			name:       "invalid pagination",
			ctx:        t.Context(),
			url:        "/accounts?limit=0",
			wantCode:   http.StatusBadRequest,
			wantDetail: "invalid pagination",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newSUT(t)
			rec := get(tt.ctx, t, repo, tt.url)

			testhelpers.AssertHumaError(t, rec, tt.wantCode, tt.wantDetail)
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
			want: page(false, false, accountID1, accountID2, accountID3),
		},
		{
			name: "set limit and offset",
			url:  "/accounts?limit=1&&offset=1",
			want: page(true, true, accountID2),
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

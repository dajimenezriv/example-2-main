package getaccounts

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/require"

	"exercise/internal/participants/app/query/getaccounts"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
	"exercise/internal/testhelpers"
)

// One more than the default limit (20), so the default is observable.
const seededAccounts = 21

var actorID = uuid.MustParse("00000000-0000-0000-0000-0000000000ff")

func accountID(n int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", n))
}

func newSUT(t *testing.T) account.Repository {
	t.Helper()
	repo := accounttest.NewFakeRepository()
	for n := 1; n <= seededAccounts; n++ {
		a, err := account.New(accountID(n), actorID, "dltID", "privKey", "pubKey")
		require.NoError(t, err)
		require.NoError(t, repo.Save(t.Context(), a))
	}
	return repo
}

func get(ctx context.Context, t *testing.T, repo account.Repository, url string) *httptest.ResponseRecorder {
	t.Helper()
	h := NewHandler(&getaccounts.Query{Repo: repo})
	api := testhelpers.NewHumaAPI(t)
	h.Register(api)
	return api.GetCtx(ctx, url)
}

// page builds the expected body for accounts [from, to]. Exact match also
// guarantees privateKey is never exposed.
func page(from, to int, hasNext, hasPrevious bool) string {
	rows := make([]string, 0, to-from+1)
	for n := from; n <= to; n++ {
		rows = append(rows, fmt.Sprintf(
			`{"id": %q, "dltAccountId": "dltID", "publicKey": "pubKey", "version": 1}`, accountID(n),
		))
	}
	return fmt.Sprintf(
		`{"accounts": [%s], "totalCount": %d, "hasNext": %t, "hasPrevious": %t}`,
		strings.Join(rows, ", "), seededAccounts, hasNext, hasPrevious,
	)
}

func TestGetAccounts_RejectedRequests(t *testing.T) {
	canceled, cancel := context.WithCancel(t.Context())
	cancel()

	tests := []struct {
		name       string
		ctx        context.Context
		url        string
		wantCode   int
		wantDetail string
		wantErrors []*huma.ErrorDetail
	}{
		{
			name:       "repository failure",
			ctx:        canceled,
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
		{
			name:       "limit above account max",
			ctx:        t.Context(),
			url:        fmt.Sprintf("/accounts?limit=%d", account.MaxPageLimit+1),
			wantCode:   http.StatusBadRequest,
			wantDetail: "invalid pagination",
		},
		{
			name:       "non-numeric limit",
			ctx:        t.Context(),
			url:        "/accounts?limit=abc",
			wantCode:   http.StatusUnprocessableEntity,
			wantDetail: "validation failed",
			wantErrors: []*huma.ErrorDetail{
				{Message: "invalid integer", Location: "query.limit", Value: "abc"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := get(tt.ctx, t, newSUT(t), tt.url)

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
			name: "default limit and offset",
			url:  "/accounts",
			want: page(1, 20, true, false),
		},
		{
			name: "limit at account max",
			url:  fmt.Sprintf("/accounts?limit=%d", account.MaxPageLimit),
			want: page(1, seededAccounts, false, false),
		},
		{
			name: "explicit limit and offset",
			url:  "/accounts?limit=2&offset=1",
			want: page(2, 3, true, true),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := get(t.Context(), t, newSUT(t), tt.url)

			testhelpers.AssertHumaOutput(t, rec, http.StatusOK, tt.want)
		})
	}
}

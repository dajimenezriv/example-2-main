package createaccount

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"exercise/internal/participants/app/command/createaccount"
	"exercise/internal/participants/app/uow"
	"exercise/internal/participants/app/uow/uowtest"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/actor/actortest"
	"exercise/internal/testhelpers"

	"github.com/danielgtaylor/huma/v2"
)

var (
	actorID   = uuid.New()
	accountID = uuid.New()
)

func newSUT(t *testing.T, newAccountID uuid.UUID) (*Handler, account.Repository) {
	t.Helper()

	act, _ := actor.New(actorID, "Tax ID", actor.Issuer)
	acc, _ := account.New(accountID, actorID, "DLT account ID", "Private key", "Public key")

	actorRepo := actortest.NewFakeRepository()
	accountRepo := accounttest.NewFakeRepository()
	_ = actorRepo.Save(t.Context(), act)
	_ = accountRepo.Save(t.Context(), acc)

	h := NewHandler(createaccount.NewHandler(
		uowtest.NewFake(uow.Repositories{Accounts: accountRepo, Actors: actorRepo}),
		func() uuid.UUID { return newAccountID },
	))
	return h, accountRepo
}

func post(t *testing.T, h *Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	api := testhelpers.NewHumaAPI(t)
	h.Register(api)
	return api.Post("/accounts", strings.NewReader(body))
}

func TestRejectedRequests(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		newAccountID uuid.UUID
		wantCode     int
		wantDetail   string
		wantErrors   []*huma.ErrorDetail
	}{
		{
			name:         "malformed body",
			body:         ``,
			newAccountID: uuid.New(),
			wantCode:     http.StatusBadRequest,
			wantDetail:   "request body is required",
		},
		{
			name:         "empty body",
			body:         `{}`,
			newAccountID: uuid.New(),
			wantCode:     http.StatusUnprocessableEntity,
			wantDetail:   "validation failed",
			wantErrors: []*huma.ErrorDetail{
				{Message: "expected required property actorId to be present", Location: "body", Value: map[string]any{}},
				{Message: "expected required property dltAccountId to be present", Location: "body", Value: map[string]any{}},
			},
		},
		{
			// Huma accepts an uppercase "URN:UUID:" prefix, but uuid.Parse does not.
			name:         "actor id rejected by uuid.Parse",
			body:         fmt.Sprintf(`{"actorId": "URN:UUID:%s", "dltAccountId": "DLT account ID"}`, uuid.New()),
			newAccountID: uuid.New(),
			wantCode:     http.StatusBadRequest,
			wantDetail:   "invalid actor id",
		},
		{
			name:         "actor not found",
			body:         fmt.Sprintf(`{"actorId": %q, "dltAccountId": "DLT account ID"}`, uuid.New()),
			newAccountID: uuid.New(),
			wantCode:     http.StatusNotFound,
			wantDetail:   "actor not found",
		},
		{
			name:         "account already exists",
			body:         fmt.Sprintf(`{"actorId": %q, "dltAccountId": "DLT account ID"}`, actorID),
			newAccountID: accountID,
			wantCode:     http.StatusConflict,
			wantDetail:   "account already exists",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, accountRepo := newSUT(t, tt.newAccountID)
			rec := post(t, h, tt.body)

			testhelpers.AssertHumaError(t, rec, tt.wantCode, tt.wantDetail, tt.wantErrors...)

			accounts, _ := accountRepo.FindAll(t.Context())
			if got := len(accounts); got != 1 {
				t.Errorf("len(accounts) = %d, want 1", got)
			}
		})
	}
}

func TestCreatesAccount(t *testing.T) {
	body := fmt.Sprintf(`{
		"actorId": %q,
		"dltAccountId": "DLT account ID",
		"publicKey": "Public key"
	}`, actorID)

	newAccountID := uuid.New()
	h, accountRepo := newSUT(t, newAccountID)
	rec := post(t, h, body)

	want := fmt.Sprintf(`{
		"id": %q,
		"actorId": %q,
		"dltAccountId": "DLT account ID",
		"publicKey": "Public key"
	}`, newAccountID, actorID)

	testhelpers.AssertHumaOutput(t, rec, http.StatusCreated, want)

	accounts, _ := accountRepo.FindAll(t.Context())
	if got := len(accounts); got != 2 {
		t.Errorf("len(accounts) = %d, want 2", got)
	}
}

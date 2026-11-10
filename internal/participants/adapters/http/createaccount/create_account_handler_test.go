package createaccount

import (
	"exercise/internal/participants/app/command/createaccount"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/actor/actortest"
	"exercise/internal/testhelpers"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"github.com/danielgtaylor/huma/v2"
)

var (
	existingActorID   = uuid.New()
	existingAccountID = uuid.New()
)

func newSUT(t *testing.T, newAccountID uuid.UUID) (*Handler, account.Repository) {
	t.Helper()

	act, _ := actor.New(existingActorID, "Tax ID", actor.Issuer)
	acc, _ := account.New(existingAccountID, existingActorID, "DLT account ID", "Private key", "Public key")

	actorRepo := actortest.NewFakeRepository()
	actorRepo.Save(t.Context(), &act)
	accountRepo := accounttest.NewFakeRepository()
	accountRepo.Save(t.Context(), &acc)

	h := NewHandler(createaccount.NewHandler(
		accountRepo,
		actorRepo,
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
				{Message: "expected required property actor_id to be present", Location: "body", Value: map[string]any{}},
				{Message: "expected required property dlt_account_id to be present", Location: "body", Value: map[string]any{}},
			},
		},
		{
			// Huma accepts an uppercase "URN:UUID:" prefix, but uuid.Parse does not.
			name:         "actor id rejected by uuid.Parse",
			body:         fmt.Sprintf(`{"actor_id": "URN:UUID:%s", "dlt_account_id": "DLT account ID"}`, uuid.New()),
			newAccountID: uuid.New(),
			wantCode:     http.StatusBadRequest,
			wantDetail:   "invalid actor id",
		},
		{
			name:         "actor not found",
			body:         fmt.Sprintf(`{"actor_id": %q, "dlt_account_id": "DLT account ID"}`, uuid.New()),
			newAccountID: uuid.New(),
			wantCode:     http.StatusNotFound,
			wantDetail:   "actor not found",
		},
		{
			name:         "account already exists",
			body:         fmt.Sprintf(`{"actor_id": %q, "dlt_account_id": "DLT account ID"}`, existingActorID),
			newAccountID: existingAccountID,
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
		"actor_id": %q,
		"dlt_account_id": "DLT account ID",
		"public_key": "Public key"
	}`, existingActorID)

	newAccountID := uuid.New()
	h, accountRepo := newSUT(t, newAccountID)
	rec := post(t, h, body)

	want := fmt.Sprintf(`{
		"id": %q,
		"actor_id": %q,
		"dlt_account_id": "DLT account ID",
		"public_key": "Public key"
	}`, newAccountID, existingActorID)

	testhelpers.AssertHumaOutput(t, rec, http.StatusCreated, want)

	accounts, _ := accountRepo.FindAll(t.Context())
	if got := len(accounts); got != 2 {
		t.Errorf("len(accounts) = %d, want 2", got)
	}
}

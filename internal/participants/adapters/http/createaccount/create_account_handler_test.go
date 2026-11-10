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
)

var (
	existingActorID   = uuid.New()
	existingAccountID = uuid.New()
)

func newSUT(t *testing.T, newAccountID uuid.UUID) (*Handler, *accounttest.FakeRepository) {
	t.Helper()

	act, err := actor.New(existingActorID, "Tax ID", actor.Issuer)
	if err != nil {
		t.Fatalf("new actor: %v", err)
	}
	acc, err := account.New(existingAccountID, existingActorID, "DLT account ID", "Private key", "Public key")
	if err != nil {
		t.Fatalf("new account: %v", err)
	}

	actorRepo := actortest.NewFakeRepository()
	if err := actorRepo.Save(t.Context(), &act); err != nil {
		t.Fatalf("save actor: %v", err)
	}
	accountRepo := accounttest.NewFakeRepository()
	if err := accountRepo.Save(t.Context(), &acc); err != nil {
		t.Fatalf("save account: %v", err)
	}

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
		},
		{
			// Huma accepts an uppercase "URN:UUID:" prefix, but uuid.Parse does not.
			name:         "actor id rejected by uuid.Parse",
			body:         `{"actor_id":"URN:UUID:` + existingActorID.String() + `","dlt_account_id":"DLT account ID"}`,
			newAccountID: uuid.New(),
			wantCode:     http.StatusBadRequest,
			wantDetail:   "invalid actor id",
		},
		{
			name:         "actor not found",
			body:         `{"actor_id":"` + uuid.New().String() + `","dlt_account_id":"DLT account ID"}`,
			newAccountID: uuid.New(),
			wantCode:     http.StatusNotFound,
			wantDetail:   "actor not found",
		},
		{
			name:         "account already exists",
			body:         `{"actor_id":"` + existingActorID.String() + `","dlt_account_id":"DLT account ID"}`,
			newAccountID: existingAccountID,
			wantCode:     http.StatusConflict,
			wantDetail:   "account already exists",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, accountRepo := newSUT(t, tt.newAccountID)
			rec := post(t, h, tt.body)

			testhelpers.AssertHumaError(t, rec, tt.wantCode, tt.wantDetail)

			if got := len(accountRepo.Accounts); got != 1 {
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

	if got := len(accountRepo.Accounts); got != 2 {
		t.Errorf("len(accounts) = %d, want 2", got)
	}
}

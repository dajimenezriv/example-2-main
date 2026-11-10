package createaccount

import (
	"exercise/internal/participants/app/command/createaccount"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/actor/actortest"
	"exercise/internal/tests"
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
	acc := account.New(existingAccountID, existingActorID, "DLT account ID", "Private key", "Public key")

	accountRepo := accounttest.NewFakeRepository()
	if err := accountRepo.Save(t.Context(), &acc); err != nil {
		t.Fatalf("seeding actor: %v", err)
	}
	actorRepo := actortest.NewFakeRepository()
	if err := actorRepo.Save(t.Context(), &act); err != nil {
		t.Fatalf("seeding actor: %v", err)
	}

	h := NewHandler(createaccount.NewHandler(
		accountRepo,
		actorRepo,
		func() uuid.UUID { return newAccountID },
	))
	return h, accountRepo
}

func post(h *Handler, body string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/accounts", strings.NewReader(body))
	h.ServeHTTP(rec, req)
	return rec
}

func TestRejectedRequests(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		newAccountID uuid.UUID
		wantCode     int
		wantBody     string
	}{
		{
			name:         "malformed body",
			body:         ``,
			newAccountID: uuid.New(),
			wantCode:     http.StatusBadRequest,
			wantBody:     "invalid request body\n",
		},
		{
			name:         "missing actor id",
			body:         `{}`,
			newAccountID: uuid.New(),
			wantCode:     http.StatusBadRequest,
			wantBody:     "invalid actor id\n",
		},
		{
			name:         "malformed actor id",
			body:         `{"actor_id": "invalid"}`,
			newAccountID: uuid.New(),
			wantCode:     http.StatusBadRequest,
			wantBody:     "invalid actor id\n",
		},
		{
			name:         "empty dlt account id",
			body:         `{"actor_id":"` + existingActorID.String() + `"}`,
			newAccountID: uuid.New(),
			wantCode:     http.StatusBadRequest,
			wantBody:     "DLT account ID is required\n",
		},
		{
			name:         "actor not found",
			body:         `{"actor_id":"` + uuid.New().String() + `","dlt_account_id":"DLT account ID"}`,
			newAccountID: uuid.New(),
			wantCode:     http.StatusNotFound,
			wantBody:     "actor not found\n",
		},
		{
			name:         "account already exists",
			body:         `{"actor_id":"` + existingActorID.String() + `","dlt_account_id":"DLT account ID"}`,
			newAccountID: existingAccountID,
			wantCode:     http.StatusConflict,
			wantBody:     "account already exists\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, accountRepo := newSUT(t, tt.newAccountID)
			rec := post(h, tt.body)

			if got := rec.Code; got != tt.wantCode {
				t.Errorf("Code = %d, want %d", got, tt.wantCode)
			}
			if got := rec.Body.String(); got != tt.wantBody {
				t.Errorf("Body = %q, want %q", got, tt.wantBody)
			}
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
	rec := post(h, body)

	if got, want := rec.Code, http.StatusCreated; got != want {
		t.Errorf("Code = %d, want %d", got, want)
	}
	if got, want := rec.Header().Get("Content-Type"), "application/json"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}

	want := fmt.Sprintf(`{
		"id": %q,
		"actor_id": %q,
		"dlt_account_id": "DLT account ID",
		"public_key": "Public key"
	}`, newAccountID, existingActorID)
	tests.AssertJSONEqual(t, rec.Body.String(), want)

	if got := len(accountRepo.Accounts); got != 2 {
		t.Errorf("len(accounts) = %d, want 2", got)
	}
}

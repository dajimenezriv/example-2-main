package createaccount

import (
	"encoding/json"
	"exercise2/cmd/actormgmt/internal/app/command/createaccount"
	"exercise2/cmd/actormgmt/internal/domain/account/accounttest"
	"exercise2/cmd/actormgmt/internal/domain/actor"
	"exercise2/cmd/actormgmt/internal/domain/actor/actortest"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"github.com/google/go-cmp/cmp"
)

var (
	existingActorID = uuid.New()
	newAccountID    = uuid.New()
)

func newSUT(t *testing.T) (*Handler, *accounttest.FakeRepository) {
	t.Helper()

	accountRepo := accounttest.NewFakeRepository()
	actorRepo := actortest.NewFakeRepository()
	actorRepo.Save(&actor.Actor{
		ID:    existingActorID,
		TaxId: "TaxID",
		Type:  actor.Issuer,
	})

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
		name     string
		body     string
		wantBody string
	}{
		{"malformed body", ``, "invalid request body\n"},
		{"missing actor id", `{}`, "invalid actor id\n"},
		{"malformed actor id", `{"actor_id":"invalid"}`, "invalid actor id\n"},
		{"create account error", `{"actor_id":"` + existingActorID.String() + `"}`, "DltAccountId is required\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, accountRepo := newSUT(t)
			rec := post(h, tt.body)

			if got, want := rec.Code, http.StatusBadRequest; got != want {
				t.Errorf("Code = %d, want %d", got, want)
			}
			if got := rec.Body.String(); got != tt.wantBody {
				t.Errorf("Body = %q, want %q", got, tt.wantBody)
			}
			if got := len(accountRepo.Accounts); got != 0 {
				t.Errorf("len(accounts) = %d, want 0", got)
			}
		})
	}
}

func TestCreateAccount(t *testing.T) {
	body := `{
		"actor_id": "` + existingActorID.String() + `",
		"dlt_account_id": "DLT account ID",
		"private_key": "Private key",
		"public_key": "Public key"
	}`

	h, accountRepo := newSUT(t)
	rec := post(h, body)

	if got, want := rec.Code, http.StatusCreated; got != want {
		t.Errorf("Code = %d, want %d", got, want)
	}
	if got, want := rec.Header().Get("Content-Type"), "application/json"; got != want {
		t.Errorf("Content-Type = %q, want %q", got, want)
	}

	var got map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("json.Decode = %v", err)
	}
	want := map[string]string{
		"id":             newAccountID.String(),
		"actor_id":       existingActorID.String(),
		"dlt_account_id": "DLT account ID",
		"private_key":    "Private key",
		"public_key":     "Public key",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("response mismatch (-want +got):\n%s", diff)
	}

	if got := len(accountRepo.Accounts); got != 1 {
		t.Errorf("len(accounts) = %d, want 1", got)
	}
}

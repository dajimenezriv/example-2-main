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
)

var (
	actorID       = uuid.New()
	nextAccountID = uuid.New()
)

func newSUT(t *testing.T, body string) (*httptest.ResponseRecorder, *http.Request) {
	t.Helper()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/accounts", strings.NewReader(body))

	actorRepo := actortest.NewFakeRepository()
	actorRepo.Save(&actor.Actor{
		ID:    actorID,
		TaxId: "TaxID",
		Type:  actor.Issuer,
	})

	h := NewHandler(createaccount.NewHandler(
		accounttest.NewFakeRepository(),
		actorRepo,
		func() uuid.UUID { return nextAccountID },
	))
	h.ServeHTTP(rec, req)
	return rec, req
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr string
	}{
		{"invalid request body", "invalid", "invalid request body\n"},
		{"invalid actor id", `{"actor_id":"invalid"}`, "invalid actor id\n"},
		{"create account error", `{"actor_id":"` + actorID.String() + `"}`, "DltAccountId is required\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, _ := newSUT(t, tt.body)
			if got, want := rec.Code, http.StatusBadRequest; got != want {
				t.Errorf("Code = %d, want %d", got, want)
			}
			if got := rec.Body.String(); got != tt.wantErr {
				t.Errorf("Body = %q, want %q", got, tt.wantErr)
			}
		})
	}
}

func TestSuccess(t *testing.T) {
	body := CreateAccountRequest{
		ActorID:      actorID.String(),
		DltAccountId: "DLT account ID",
		PrivateKey:   "Private key",
		PublicKey:    "Public key",
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("json.Marshal = %v", err)
	}

	rec, _ := newSUT(t, string(b))
	if got, want := rec.Code, http.StatusCreated; got != want {
		t.Errorf("Code = %d, want %d", got, want)
	}
	var got CreateAccountResponse
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("json.Decode = %v", err)
	}
	want := CreateAccountResponse{
		ID:           nextAccountID.String(),
		ActorID:      body.ActorID,
		DltAccountId: body.DltAccountId,
		PrivateKey:   body.PrivateKey,
		PublicKey:    body.PublicKey,
	}
	if got != want {
		t.Errorf("Response:\n got  = %+v\n want = %+v", got, want)
	}
}

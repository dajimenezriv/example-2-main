package createaccount

import (
	"errors"
	"exercise/internal/domain/account"
	"exercise/internal/domain/account/accounttest"
	"exercise/internal/domain/actor"
	"exercise/internal/domain/actor/actortest"
	"testing"

	"uuid"

	"github.com/google/go-cmp/cmp"
)

var (
	exitingActorID   = uuid.New()
	exitingAccountID = uuid.New()
)

func newSUT(t *testing.T, newAccountID uuid.UUID) (*Handler, *accounttest.FakeRepository, *actortest.FakeRepository) {
	t.Helper()

	act, err := actor.New(exitingActorID, "Tax ID", actor.Issuer)
	if err != nil {
		t.Fatalf("actor.New = %v", err)
	}
	acc := account.New(exitingAccountID, exitingActorID, "DLT account ID", "Private key", "Public key")

	accRepo := accounttest.NewFakeRepository()
	if err := accRepo.Save(t.Context(), &acc); err != nil {
		t.Fatalf("seeding actor: %v", err)
	}
	actRepo := actortest.NewFakeRepository()
	if err := actRepo.Save(t.Context(), &act); err != nil {
		t.Fatalf("seeding actor: %v", err)
	}

	newID := func() uuid.UUID { return newAccountID }
	return NewHandler(accRepo, actRepo, newID), accRepo, actRepo
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name    string
		cmd     Command
		wantErr error
	}{
		{
			name:    "empty dlt account id",
			cmd:     Command{},
			wantErr: ErrDltAccountIdRequired,
		},
		{
			name:    "actor not found",
			cmd:     Command{ActorID: uuid.UUID{}, DltAccountId: "DLT account ID"},
			wantErr: actor.ErrActorNotFound,
		},
		{
			name:    "account already exists",
			cmd:     Command{ActorID: exitingActorID, DltAccountId: "DLT account ID"},
			wantErr: account.ErrAccountAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, _, _ := newSUT(t, exitingAccountID)

			got, err := h.Execute(t.Context(), tt.cmd)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Execute() = %v, want %v", err, tt.wantErr)
			}
			if got != nil {
				t.Errorf("Execute() = %+v, want nil", got)
			}
		})
	}
}

func TestCreatesAccount(t *testing.T) {
	newAccountID := uuid.New()
	h, accRepo, _ := newSUT(t, newAccountID)
	cmd := Command{
		ActorID:      exitingActorID,
		DltAccountId: "DLT account ID",
		PrivateKey:   "Private key",
		PublicKey:    "Public key",
	}

	acc, err := h.Execute(t.Context(), cmd)

	if err != nil {
		t.Fatalf("Execute() = %v, want nil", err)
	}

	want := &account.Account{
		ID:           newAccountID,
		ActorID:      exitingActorID,
		DltAccountId: cmd.DltAccountId,
		PrivateKey:   cmd.PrivateKey,
		PublicKey:    cmd.PublicKey,
	}
	if diff := cmp.Diff(want, acc); diff != "" {
		t.Errorf("Execute() account mismatch (-want +got):\n%s", diff)
	}

	stored, err := accRepo.FindByID(t.Context(), newAccountID)
	if err != nil {
		t.Fatalf("FindByID(%v) = %v, want stored account", newAccountID, err)
	}
	if diff := cmp.Diff(want, stored); diff != "" {
		t.Errorf("stored account mismatch (-want +got):\n%s", diff)
	}
	if got := len(accRepo.Accounts); got != 2 {
		t.Errorf("len(accounts) = %d, want 2", got)
	}
}

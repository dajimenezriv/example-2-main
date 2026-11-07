package createaccount

import (
	"errors"
	"exercise2/cmd/actormgmt/internal/domain/account"
	"exercise2/cmd/actormgmt/internal/domain/account/accounttest"
	"exercise2/cmd/actormgmt/internal/domain/actor"
	"exercise2/cmd/actormgmt/internal/domain/actor/actortest"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
)

var (
	errBoom        = errors.New("boom")
	exitingActorID = uuid.New()
)

func newSUT(t *testing.T) (*Handler, *accounttest.FakeRepository, *actortest.FakeRepository) {
	t.Helper()
	accRepo := accounttest.NewFakeRepository()
	actRepo := actortest.NewFakeRepository()
	if err := actRepo.Save(&actor.Actor{ID: exitingActorID}); err != nil {
		t.Fatalf("seeding actor: %v", err)
	}
	return NewHandler(accRepo, actRepo), accRepo, actRepo
}

func TestErrors(t *testing.T) {
	tests := []struct {
		name    string
		cmd     Command
		saveErr error
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
			name:    "account save error",
			cmd:     Command{ActorID: exitingActorID, DltAccountId: "DLT account ID"},
			saveErr: errBoom,
			wantErr: errBoom,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, accRepo, _ := newSUT(t)

			accRepo.SaveErr = tt.saveErr
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
	h, accRepo, _ := newSUT(t)
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
		ActorID:      exitingActorID,
		DltAccountId: cmd.DltAccountId,
		PrivateKey:   cmd.PrivateKey,
		PublicKey:    cmd.PublicKey,
	}
	if diff := cmp.Diff(want, acc, cmpopts.IgnoreFields(account.Account{}, "ID")); diff != "" {
		t.Errorf("Execute() account mismatch (-want +got):\n%s", diff)
	}
	if acc.ID == uuid.Nil {
		t.Error("Execute() account ID was not generated")
	}

	// The account was persisted exactly as returned. Read it back through the
	// port rather than the fake's internals.
	stored, err := accRepo.FindByID(acc.ID)
	if err != nil {
		t.Fatalf("FindByID(%v) = %v, want stored account", acc.ID, err)
	}
	if diff := cmp.Diff(acc, stored); diff != "" {
		t.Errorf("stored account mismatch (-returned +stored):\n%s", diff)
	}
	if got := len(accRepo.Accounts); got != 1 {
		t.Errorf("len(accounts) = %d, want 1", got)
	}
}

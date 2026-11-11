package account

import (
	"errors"
	"testing"
	"uuid"

	"github.com/google/go-cmp/cmp"
)

func TestNewAccountErrors(t *testing.T) {
	a, err := New(uuid.New(), uuid.New(), "", "", "")
	if !errors.Is(err, ErrDltAccountIDRequired) {
		t.Errorf("New() = %v, want %v", err, ErrDltAccountIDRequired)
	}
	if diff := cmp.Diff(a, Account{}); diff != "" {
		t.Errorf("account (-want +got):\n%s", diff)
	}
}

func TestNewCreatesValidAccount(t *testing.T) {
	id, actorID := uuid.New(), uuid.New()
	a, err := New(id, actorID, "DLT account ID", "Private key", "Public key")
	if err != nil {
		t.Errorf("New() = %v, want nil", err)
	}
	want := Account{
		ID:           id,
		ActorID:      actorID,
		DltAccountID: "DLT account ID",
		PrivateKey:   "Private key",
		PublicKey:    "Public key",
	}
	if diff := cmp.Diff(a, want); diff != "" {
		t.Errorf("actor (-want +got):\n%s", diff)
	}
}

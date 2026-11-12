package account_test

import (
	"errors"
	"testing"
	"uuid"

	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
)

func TestNewAccountErrors(t *testing.T) {
	got, err := account.New(uuid.New(), uuid.New(), "", "", "")
	if !errors.Is(err, account.ErrDltAccountIDRequired) {
		t.Errorf("New() = %v, want %v", err, account.ErrDltAccountIDRequired)
	}
	accounttest.AssertAccount(t, got, nil)
}

func TestNewCreatesValidAccount(t *testing.T) {
	id, actorID := uuid.New(), uuid.New()
	got, err := account.New(id, actorID, "DLT account ID", "Private key", "Public key")
	if err != nil {
		t.Errorf("New() = %v, want nil", err)
	}
	want := account.Reconstitute(id, actorID, "DLT account ID", "Private key", "Public key")
	accounttest.AssertAccount(t, got, want)
}

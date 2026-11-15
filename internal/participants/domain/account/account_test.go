package account_test

import (
	"errors"
	"testing"
	"uuid"

	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
)

func TestNewAccountErrors(t *testing.T) {
	tests := []struct {
		name    string
		dltID   string
		privKey string
		pubKey  string
		wantErr error
	}{
		{"empty DLT account ID", "", "privKey", "pubKey", account.ErrDltAccountIDRequired},
		{"empty private key", "dltID", "", "pubKey", account.ErrPrivateKeyRequired},
		{"empty public key", "dltID", "privKey", "", account.ErrPublicKeyRequired},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := account.New(uuid.New(), uuid.New(), tt.dltID, tt.privKey, tt.pubKey)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("New() = %v, want %v", err, tt.wantErr)
			}
			if got != nil {
				t.Errorf("New() = %v, want nil", got)
			}
		})
	}
}

func TestNewCreatesValidAccount(t *testing.T) {
	id, actorID := uuid.New(), uuid.New()
	got, err := account.New(id, actorID, "dltID", "privKey", "pubKey")
	if err != nil {
		t.Errorf("New() = %v, want nil", err)
	}
	want := account.Reconstitute(id, actorID, "dltID", "privKey", "pubKey", 1)
	accounttest.AssertAccount(t, got, want)
}

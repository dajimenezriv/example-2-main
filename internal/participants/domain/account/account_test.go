package account_test

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"exercise/internal/participants/domain/account"
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
			require.ErrorIs(t, err, tt.wantErr)
			assert.Nil(t, got)
		})
	}
}

func TestNewCreatesValidAccount(t *testing.T) {
	id, actorID := uuid.New(), uuid.New()
	got, err := account.New(id, actorID, "dltID", "privKey", "pubKey")
	require.NoError(t, err)
	want := account.Reconstitute(id, actorID, "dltID", "privKey", "pubKey", 1)
	assert.Equal(t, want, got)
}

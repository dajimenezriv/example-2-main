package actor_test

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"exercise/internal/participants/domain/actor"
)

func TestNewActorErrors(t *testing.T) {
	tests := []struct {
		name    string
		taxID   string
		t       actor.Type
		wantErr error
	}{
		{"invalid tax ID", "invalid", actor.Issuer, actor.ErrInvalidTaxID},
		{"invalid actor type", "123456789", "invalid", actor.ErrInvalidType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := actor.New(uuid.New(), tt.taxID, tt.t)
			require.ErrorIs(t, err, tt.wantErr)
			assert.Nil(t, got)
		})
	}
}

func TestNewCreatesValidActor(t *testing.T) {
	id := uuid.New()
	got, err := actor.New(id, "123456789", actor.Issuer)
	require.NoError(t, err)
	want := actor.Reconstitute(id, "123456789", actor.Issuer, 1)
	assert.Equal(t, want, got)
}

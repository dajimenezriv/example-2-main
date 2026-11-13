package actor_test

import (
	"errors"
	"testing"
	"uuid"

	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/actor/actortest"
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
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("New() = %v, want %v", err, tt.wantErr)
			}
			if got != nil {
				t.Errorf("New() = %v, want nil", got)
			}
		})
	}
}

func TestNewCreatesValidActor(t *testing.T) {
	id := uuid.New()
	got, err := actor.New(id, "123456789", actor.Issuer)
	if err != nil {
		t.Errorf("New() = %v, want nil", err)
	}
	want := actor.Reconstitute(id, "123456789", actor.Issuer, 1)
	actortest.AssertActor(t, got, want)
}

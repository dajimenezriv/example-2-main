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
		{"empty taxID", "", actor.Issuer, actor.ErrTaxIDRequired},
		{"invalid actor type", "Tax ID", "invalid", actor.ErrInvalidActorType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := actor.New(uuid.New(), tt.taxID, tt.t)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("New() = %v, want %v", err, tt.wantErr)
			}
			actortest.AssertActor(t, got, nil)
		})
	}
}

func TestNewCreatesValidActor(t *testing.T) {
	id := uuid.New()
	got, err := actor.New(id, "Tax ID", actor.Issuer)
	if err != nil {
		t.Errorf("New() = %v, want nil", err)
	}
	want := actor.Reconstitute(id, "Tax ID", actor.Issuer)
	actortest.AssertActor(t, got, want)
}

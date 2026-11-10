package actor

import (
	"errors"
	"testing"
	"uuid"

	"github.com/google/go-cmp/cmp"
)

func TestNewActorErrors(t *testing.T) {
	tests := []struct {
		name    string
		taxID   string
		t       ActorType
		wantErr error
	}{
		{"empty taxID", "", Issuer, ErrTaxIDRequired},
		{"invalid actor type", "Tax ID", "invalid", ErrInvalidActorType},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, err := New(uuid.New(), tt.taxID, tt.t)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("New() = %v, want %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(a, Actor{}); diff != "" {
				t.Errorf("actor (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNewCreatesValidActor(t *testing.T) {
	id := uuid.New()
	a, err := New(id, "Tax ID", Issuer)
	if err != nil {
		t.Errorf("New() = %v, want nil", err)
	}
	want := Actor{ID: id, TaxId: "Tax ID", Type: Issuer}
	if diff := cmp.Diff(a, want); diff != "" {
		t.Errorf("actor (-want +got):\n%s", diff)
	}
}

package actor

import (
	"errors"
	"testing"
	"uuid"

	"github.com/google/go-cmp/cmp"
)

func TestNewInvalidActorType(t *testing.T) {
	a, err := New(uuid.New(), "TaxID", "invalid")
	if !errors.Is(err, ErrInvalidActorType) {
		t.Errorf("New() = %v, want %v", err, ErrInvalidActorType)
	}
	if diff := cmp.Diff(a, Actor{}); diff != "" {
		t.Errorf("actor (-want +got):\n%s", diff)
	}
}

func TestNewCreatesValidActor(t *testing.T) {
	id := uuid.New()
	a, err := New(id, "TaxID", Issuer)
	if err != nil {
		t.Errorf("New() = %v, want nil", err)
	}
	want := Actor{ID: id, TaxId: "TaxID", Type: Issuer}
	if diff := cmp.Diff(a, want); diff != "" {
		t.Errorf("actor (-want +got):\n%s", diff)
	}
}

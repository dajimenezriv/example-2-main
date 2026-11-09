package actortest

import (
	"errors"
	"exercise2/cmd/actormgmt/internal/domain/actor"
	"testing"
	"uuid"

	"github.com/google/go-cmp/cmp"
)

func RunRepositoryContract(t *testing.T, newRepo func(t *testing.T) actor.Repository) {
	t.Run("Save stores the actor", func(t *testing.T) {
		repo := newRepo(t)
		want, _ := actor.New("TaxID", actor.Issuer)

		if err := repo.Save(&want); err != nil {
			t.Fatalf("Save() = %v, want nil", err)
		}

		got, err := repo.FindByID(want.ID)
		if err != nil {
			t.Errorf("FindByID() = %v, want nil", err)
		}
		if diff := cmp.Diff(&want, got); diff != "" {
			t.Errorf("actor mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("Save fails if actor already exists", func(t *testing.T) {
		repo := newRepo(t)
		want, _ := actor.New("TaxID", actor.Issuer)

		if err := repo.Save(&want); err != nil {
			t.Fatalf("Save() = %v, want nil", err)
		}

		duplicate := want
		duplicate.TaxId = "OtherTaxID"
		if err := repo.Save(&duplicate); !errors.Is(err, actor.ErrActorAlreadyExists) {
			t.Fatalf("Save() = %v, want %v", err, actor.ErrActorAlreadyExists)
		}

		got, err := repo.FindByID(want.ID)
		if err != nil {
			t.Errorf("FindByID() = %v, want nil", err)
		}
		if diff := cmp.Diff(&want, got); diff != "" {
			t.Errorf("actor mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("FindByID returns error if actor not found", func(t *testing.T) {
		repo := newRepo(t)
		a, err := repo.FindByID(uuid.New())

		if !errors.Is(err, actor.ErrActorNotFound) {
			t.Errorf("FindByID() = %v, want %v", err, actor.ErrActorNotFound)
		}
		if a != nil {
			t.Errorf("FindByID() = %+v, want nil", a)
		}
	})
}

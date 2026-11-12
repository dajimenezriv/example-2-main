package actortest

import (
	"errors"
	"testing"
	"uuid"

	"exercise/internal/participants/domain/actor"
)

func newActor(t *testing.T, taxID string) *actor.Actor {
	t.Helper()
	a, _ := actor.New(uuid.New(), taxID, actor.Issuer)
	return a
}

func saveActor(t *testing.T, repo actor.Repository, taxID string) *actor.Actor {
	t.Helper()
	a := newActor(t, taxID)
	if err := repo.Save(t.Context(), a); err != nil {
		t.Fatalf("save actor: %v", err)
	}
	return a
}

func RunRepositoryContract(t *testing.T, newRepo func(t *testing.T) actor.Repository) {
	t.Helper()

	t.Run("Save stores the actor", func(t *testing.T) {
		repo := newRepo(t)
		want := newActor(t, "123456789")

		if err := repo.Save(t.Context(), want); err != nil {
			t.Fatalf("Save() = %v, want nil", err)
		}

		got, err := repo.FindByID(t.Context(), want.ID())
		if err != nil {
			t.Errorf("FindByID() = %v, want nil", err)
		}
		AssertActor(t, got, want)
	})

	t.Run("Save fails if ID already exists", func(t *testing.T) {
		repo := newRepo(t)
		want := newActor(t, "000000000")

		if err := repo.Save(t.Context(), want); err != nil {
			t.Fatalf("Save() = %v, want nil", err)
		}

		duplicate, _ := actor.New(want.ID(), "999999999", want.Type())
		if err := repo.Save(t.Context(), duplicate); !errors.Is(err, actor.ErrAlreadyExists) {
			t.Fatalf("Save() = %v, want %v", err, actor.ErrAlreadyExists)
		}

		got, err := repo.FindByID(t.Context(), want.ID())
		if err != nil {
			t.Errorf("FindByID() = %v, want nil", err)
		}
		AssertActor(t, got, want)
	})

	t.Run("Save fails if tax ID already exists", func(t *testing.T) {
		repo := newRepo(t)
		want := newActor(t, "000000000")

		if err := repo.Save(t.Context(), want); err != nil {
			t.Fatalf("Save() = %v, want nil", err)
		}

		duplicate, _ := actor.New(uuid.New(), string(want.TaxID()), want.Type())
		if err := repo.Save(t.Context(), duplicate); !errors.Is(err, actor.ErrAlreadyExists) {
			t.Fatalf("Save() = %v, want %v", err, actor.ErrAlreadyExists)
		}

		got, err := repo.FindByID(t.Context(), want.ID())
		if err != nil {
			t.Errorf("FindByID() = %v, want nil", err)
		}
		AssertActor(t, got, want)
	})

	t.Run("FindByID returns error if actor not found", func(t *testing.T) {
		repo := newRepo(t)
		a, err := repo.FindByID(t.Context(), uuid.New())

		if !errors.Is(err, actor.ErrNotFound) {
			t.Errorf("FindByID() = %v, want %v", err, actor.ErrNotFound)
		}
		if a != nil {
			t.Errorf("FindByID() = %+v, want nil", a)
		}
	})

	t.Run("FindAll returns all actors in inserted order", func(t *testing.T) {
		repo := newRepo(t)
		want := []*actor.Actor{
			saveActor(t, repo, "000000000"),
			saveActor(t, repo, "000000001"),
			saveActor(t, repo, "000000002"),
			saveActor(t, repo, "000000003"),
		}

		got, err := repo.FindAll(t.Context())
		if err != nil {
			t.Errorf("FindAll() = %v, want nil", err)
		}
		AssertActor(t, got, want)
	})
}

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
		want := saveActor(t, repo, "000000000")
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
		want := saveActor(t, repo, "000000000")
		duplicate := newActor(t, string(want.TaxID()))

		if err := repo.Save(t.Context(), duplicate); !errors.Is(err, actor.ErrTaxIDAlreadyExists) {
			t.Fatalf("Save() = %v, want %v", err, actor.ErrTaxIDAlreadyExists)
		}

		got, err := repo.FindByID(t.Context(), want.ID())
		if err != nil {
			t.Errorf("FindByID() = %+v, want nil", err)
		}
		AssertActor(t, got, want)
	})

	t.Run("Update modifies the actor and increases version", func(t *testing.T) {
		repo := newRepo(t)
		a := saveActor(t, repo, "000000000")
		a.ChangeTaxID("999999999")

		got, err := repo.Update(t.Context(), a)
		if err != nil {
			t.Errorf("Update() = %v, want nil", err)
		}
		want := actor.Reconstitute(a.ID(), string(a.TaxID()), a.Type(), 2)
		AssertActor(t, got, want)
	})

	t.Run("Update fails if actor does not exist", func(t *testing.T) {
		repo := newRepo(t)
		got, err := repo.Update(t.Context(), newActor(t, "123456789"))

		if !errors.Is(err, actor.ErrNotFound) {
			t.Errorf("Update() = %v, want %v", err, actor.ErrNotFound)
		}
		if got != nil {
			t.Errorf("Update() = %+v, want nil", got)
		}
	})

	t.Run("Update fails if tax ID already exists", func(t *testing.T) {
		repo := newRepo(t)
		a1 := saveActor(t, repo, "000000000")
		a2 := saveActor(t, repo, "999999999")
		a1.ChangeTaxID(string(a2.TaxID()))

		got, err := repo.Update(t.Context(), a1)
		if !errors.Is(err, actor.ErrTaxIDAlreadyExists) {
			t.Errorf("Update() = %v, want %v", err, actor.ErrTaxIDAlreadyExists)
		}
		if got != nil {
			t.Fatalf("Update() = %+v, want nil", got)
		}

		got, err = repo.FindByID(t.Context(), a1.ID())
		if err != nil {
			t.Errorf("FindByID() = %+v, want nil", err)
		}
		a1.ChangeTaxID("000000000")
		AssertActor(t, got, a1)
	})

	t.Run("Update fails if version does not match", func(t *testing.T) {
		repo := newRepo(t)
		got, err := repo.Update(t.Context(), newActor(t, "123456789"))

		if !errors.Is(err, actor.ErrNotFound) {
			t.Errorf("Update() = %v, want %v", err, actor.ErrNotFound)
		}
		if got != nil {
			t.Errorf("Update() = %+v, want nil", got)
		}
	})

	t.Run("FindByID returns actor if exists", func(t *testing.T) {
		repo := newRepo(t)
		want := saveActor(t, repo, "123456789")

		got, err := repo.FindByID(t.Context(), want.ID())
		if err != nil {
			t.Errorf("FindByID) = %v, want nil", err)
		}
		AssertActor(t, got, want)
	})

	t.Run("FindByID returns error if actor not found", func(t *testing.T) {
		repo := newRepo(t)
		got, err := repo.FindByID(t.Context(), uuid.New())

		if !errors.Is(err, actor.ErrNotFound) {
			t.Errorf("FindByID() = %v, want %v", err, actor.ErrNotFound)
		}
		if got != nil {
			t.Errorf("FindByID() = %+v, want nil", got)
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

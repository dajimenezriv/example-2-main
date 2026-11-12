package actortest

import (
	"errors"
	"testing"
	"uuid"

	"exercise/internal/participants/domain/actor"
)

func newActor(t *testing.T) *actor.Actor {
	t.Helper()
	a, _ := actor.New(uuid.New(), "Tax ID", actor.Issuer)
	return a
}

func saveActor(t *testing.T, repo actor.Repository) *actor.Actor {
	t.Helper()
	a := newActor(t)
	_ = repo.Save(t.Context(), a)
	return a
}

func RunRepositoryContract(t *testing.T, newRepo func(t *testing.T) actor.Repository) {
	t.Helper()

	t.Run("Save stores the actor", func(t *testing.T) {
		repo := newRepo(t)
		want := newActor(t)

		if err := repo.Save(t.Context(), want); err != nil {
			t.Fatalf("Save() = %v, want nil", err)
		}

		got, err := repo.FindByID(t.Context(), want.ID())
		if err != nil {
			t.Errorf("FindByID() = %v, want nil", err)
		}
		AssertActor(t, got, want)
	})

	t.Run("Save fails if actor already exists", func(t *testing.T) {
		repo := newRepo(t)
		want := newActor(t)

		if err := repo.Save(t.Context(), want); err != nil {
			t.Fatalf("Save() = %v, want nil", err)
		}

		duplicate, _ := actor.New(want.ID(), "Other Tax ID", want.Type())
		if err := repo.Save(t.Context(), duplicate); !errors.Is(err, actor.ErrActorAlreadyExists) {
			t.Fatalf("Save() = %v, want %v", err, actor.ErrActorAlreadyExists)
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

		if !errors.Is(err, actor.ErrActorNotFound) {
			t.Errorf("FindByID() = %v, want %v", err, actor.ErrActorNotFound)
		}
		if a != nil {
			t.Errorf("FindByID() = %+v, want nil", a)
		}
	})

	t.Run("FindAll returns all actors in inserted order", func(t *testing.T) {
		repo := newRepo(t)
		want := []*actor.Actor{
			saveActor(t, repo),
			saveActor(t, repo),
			saveActor(t, repo),
			saveActor(t, repo),
		}

		got, err := repo.FindAll(t.Context())
		if err != nil {
			t.Errorf("FindAll() = %v, want nil", err)
		}
		AssertActor(t, got, want)
	})
}

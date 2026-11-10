package actortest

import (
	"errors"
	"exercise/internal/participants/domain/actor"
	"testing"
	"uuid"

	"github.com/google/go-cmp/cmp"
)

func newActor(t *testing.T) actor.Actor {
	t.Helper()
	a, err := actor.New(uuid.New(), "Tax ID", actor.Issuer)
	if err != nil {
		t.Fatalf("actor.New = %v", err)
	}
	return a
}

func saveActor(t *testing.T, repo actor.Repository) *actor.Actor {
	t.Helper()
	a := newActor(t)
	if err := repo.Save(t.Context(), &a); err != nil {
		t.Fatalf("repo.Save = %v", err)
	}
	return &a
}

func RunRepositoryContract(t *testing.T, newRepo func(t *testing.T) actor.Repository) {
	t.Run("Save stores the actor", func(t *testing.T) {
		repo := newRepo(t)
		want := newActor(t)

		if err := repo.Save(t.Context(), &want); err != nil {
			t.Fatalf("Save() = %v, want nil", err)
		}

		got, err := repo.FindByID(t.Context(), want.ID)
		if err != nil {
			t.Errorf("FindByID() = %v, want nil", err)
		}
		if diff := cmp.Diff(&want, got); diff != "" {
			t.Errorf("actor (-want +got):\n%s", diff)
		}
	})

	t.Run("Save fails if actor already exists", func(t *testing.T) {
		repo := newRepo(t)
		want := newActor(t)

		if err := repo.Save(t.Context(), &want); err != nil {
			t.Fatalf("Save() = %v, want nil", err)
		}

		duplicate := want
		duplicate.TaxID = "Other Tax ID"
		if err := repo.Save(t.Context(), &duplicate); !errors.Is(err, actor.ErrActorAlreadyExists) {
			t.Fatalf("Save() = %v, want %v", err, actor.ErrActorAlreadyExists)
		}

		got, err := repo.FindByID(t.Context(), want.ID)
		if err != nil {
			t.Errorf("FindByID() = %v, want nil", err)
		}
		if diff := cmp.Diff(&want, got); diff != "" {
			t.Errorf("actor (-want +got):\n%s", diff)
		}
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
		want := make([]*actor.Actor, 0, 4)
		want = append(want, saveActor(t, repo))
		want = append(want, saveActor(t, repo))
		want = append(want, saveActor(t, repo))
		want = append(want, saveActor(t, repo))

		got, err := repo.FindAll(t.Context())
		if err != nil {
			t.Errorf("FindAll() = %v, want nil", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("actors (-want +got):\n%s", diff)
		}
	})
}

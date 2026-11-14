package actortest

import (
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"exercise/internal/participants/domain/actor"
)

func newActor(t *testing.T, taxID string) *actor.Actor {
	t.Helper()
	a, err := actor.New(uuid.New(), taxID, actor.Issuer)
	require.NoError(t, err)
	return a
}

func saveActor(t *testing.T, repo actor.Repository, taxID string) *actor.Actor {
	t.Helper()
	a := newActor(t, taxID)
	err := repo.Save(t.Context(), a)
	require.NoError(t, err)
	return a
}

func checkActor(t *testing.T, repo actor.Repository, want *actor.Actor) {
	t.Helper()
	got, err := repo.FindByID(t.Context(), want.ID())
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func RunRepositoryContract(t *testing.T, newRepo func(t *testing.T) actor.Repository) {
	t.Helper()

	t.Run("Save stores the actor", func(t *testing.T) {
		repo := newRepo(t)
		a := newActor(t, "000000000")

		err := repo.Save(t.Context(), a)
		require.NoError(t, err)
		checkActor(t, repo, a)
	})

	t.Run("Save fails if ID already exists", func(t *testing.T) {
		repo := newRepo(t)
		a := saveActor(t, repo, "000000000")
		duplicate, err := actor.New(a.ID(), "999999999", a.Type())
		require.NoError(t, err)

		err = repo.Save(t.Context(), duplicate)
		require.ErrorIs(t, err, actor.ErrAlreadyExists)
		checkActor(t, repo, a)
	})

	t.Run("Save fails if tax ID already exists", func(t *testing.T) {
		repo := newRepo(t)
		saveActor(t, repo, "000000000")
		a := newActor(t, "000000000")

		err := repo.Save(t.Context(), a)
		require.ErrorIs(t, err, actor.ErrTaxIDAlreadyExists)

		actors, err := repo.FindAll(t.Context())
		assert.NoError(t, err)
		assert.Len(t, actors, 1)
	})

	t.Run("Update modifies the actor and increases version", func(t *testing.T) {
		repo := newRepo(t)
		a := saveActor(t, repo, "000000000")
		err := a.ChangeTaxID("999999999")
		require.NoError(t, err)

		got, err := repo.Update(t.Context(), a, 1)
		require.NoError(t, err)
		want := actor.Reconstitute(a.ID(), "999999999", a.Type(), 2)
		assert.Equal(t, want, got)
	})

	t.Run("Update without modifying the actor increases version", func(t *testing.T) {
		repo := newRepo(t)
		a := saveActor(t, repo, "000000000")

		got, err := repo.Update(t.Context(), a, 1)
		require.NoError(t, err)
		want := actor.Reconstitute(a.ID(), "000000000", a.Type(), 2)
		assert.Equal(t, want, got)
	})

	t.Run("Update fails if actor does not exist", func(t *testing.T) {
		repo := newRepo(t)
		a := newActor(t, "000000000")

		got, err := repo.Update(t.Context(), a, 1)
		assert.ErrorIs(t, err, actor.ErrNotFound)
		assert.Nil(t, got)
	})

	t.Run("Update fails if tax ID already exists", func(t *testing.T) {
		repo := newRepo(t)
		a := saveActor(t, repo, "000000000")
		saveActor(t, repo, "999999999")

		err := a.ChangeTaxID("999999999")
		require.NoError(t, err)
		got, err := repo.Update(t.Context(), a, 1)
		assert.ErrorIs(t, err, actor.ErrTaxIDAlreadyExists)
		assert.Nil(t, got)

		want := actor.Reconstitute(a.ID(), "000000000", a.Type(), 1)
		checkActor(t, repo, want)
	})

	t.Run("Update fails if versions do not match", func(t *testing.T) {
		repo := newRepo(t)
		a := saveActor(t, repo, "000000000")

		err := a.ChangeTaxID("999999999")
		require.NoError(t, err)
		got, err := repo.Update(t.Context(), a, 2)
		assert.ErrorIs(t, err, actor.ErrVersionConflict)
		assert.Nil(t, got)

		want := actor.Reconstitute(a.ID(), "000000000", a.Type(), 1)
		checkActor(t, repo, want)
	})

	t.Run("Update checks if exists before tax ID uniqueness", func(t *testing.T) {
		repo := newRepo(t)
		saveActor(t, repo, "000000000")
		a := newActor(t, "999999999")

		err := a.ChangeTaxID("000000000")
		require.NoError(t, err)
		got, err := repo.Update(t.Context(), a, 1)
		assert.ErrorIs(t, err, actor.ErrNotFound)
		assert.Nil(t, got)
	})

	t.Run("Update checks version before tax ID uniqueness", func(t *testing.T) {
		repo := newRepo(t)
		saveActor(t, repo, "000000000")
		a := saveActor(t, repo, "999999999")

		err := a.ChangeTaxID("000000000")
		require.NoError(t, err)
		got, err := repo.Update(t.Context(), a, 2)
		assert.ErrorIs(t, err, actor.ErrVersionConflict)
		assert.Nil(t, got)

		want := actor.Reconstitute(a.ID(), "999999999", a.Type(), 1)
		checkActor(t, repo, want)
	})

	t.Run("FindByID returns actor if exists", func(t *testing.T) {
		repo := newRepo(t)
		a := saveActor(t, repo, "000000000")

		got, err := repo.FindByID(t.Context(), a.ID())
		assert.NoError(t, err)
		assert.Equal(t, a, got)
	})

	t.Run("FindByID returns error if actor not found", func(t *testing.T) {
		repo := newRepo(t)

		got, err := repo.FindByID(t.Context(), uuid.New())
		assert.ErrorIs(t, err, actor.ErrNotFound)
		assert.Nil(t, got)
	})

	t.Run("FindAll returns all actors in inserted order", func(t *testing.T) {
		repo := newRepo(t)
		actors := []*actor.Actor{
			saveActor(t, repo, "000000000"),
			saveActor(t, repo, "111111111"),
			saveActor(t, repo, "222222222"),
			saveActor(t, repo, "444444444"),
		}

		got, err := repo.FindAll(t.Context())
		assert.NoError(t, err)
		assert.Equal(t, actors, got)
	})
}

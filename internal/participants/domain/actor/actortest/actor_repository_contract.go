package actortest

import (
	"context"
	"math"
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/pagination"
)

func newActor(t *testing.T, taxID string) *actor.Actor {
	t.Helper()
	a, err := actor.New(uuid.NewV7(), taxID, actor.Issuer)
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

func newPage(t *testing.T, limit, offset int32) pagination.Page {
	t.Helper()
	p, err := pagination.NewPage(limit, offset, actor.MaxPageLimit)
	require.NoError(t, err)
	return p
}

func checkActor(t *testing.T, repo actor.Repository, want *actor.Actor) {
	t.Helper()
	got, err := repo.FindByID(t.Context(), want.ID())
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func RunRepositoryContract(t *testing.T, newRepo func(t *testing.T) actor.Repository) {
	t.Helper()
	runCountContract(t, newRepo)
	runSaveContract(t, newRepo)
	runUpdateContract(t, newRepo)
	runDeleteContract(t, newRepo)
	runFindContract(t, newRepo)
	runContextCanceledContract(t, newRepo)
}

func runCountContract(t *testing.T, newRepo func(t *testing.T) actor.Repository) {
	t.Helper()

	t.Run("Count returns zero when there are no actors", func(t *testing.T) {
		repo := newRepo(t)

		got, err := repo.Count(t.Context())
		require.NoError(t, err)
		assert.Equal(t, int64(0), got)
	})

	t.Run("Count returns the number of stored actors", func(t *testing.T) {
		repo := newRepo(t)
		saveActor(t, repo, "000000000")
		saveActor(t, repo, "111111111")
		saveActor(t, repo, "222222222")

		got, err := repo.Count(t.Context())
		require.NoError(t, err)
		assert.Equal(t, int64(3), got)
	})
}

func runSaveContract(t *testing.T, newRepo func(t *testing.T) actor.Repository) {
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

		count, err := repo.Count(t.Context())
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}

func runUpdateContract(t *testing.T, newRepo func(t *testing.T) actor.Repository) {
	t.Helper()

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
		require.ErrorIs(t, err, actor.ErrNotFound)
		assert.Nil(t, got)
	})

	t.Run("Update fails if tax ID already exists", func(t *testing.T) {
		repo := newRepo(t)
		a := saveActor(t, repo, "000000000")
		saveActor(t, repo, "999999999")

		err := a.ChangeTaxID("999999999")
		require.NoError(t, err)
		got, err := repo.Update(t.Context(), a, 1)
		require.ErrorIs(t, err, actor.ErrTaxIDAlreadyExists)
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
		require.ErrorIs(t, err, actor.ErrVersionConflict)
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
		require.ErrorIs(t, err, actor.ErrNotFound)
		assert.Nil(t, got)
	})

	t.Run("Update checks version before tax ID uniqueness", func(t *testing.T) {
		repo := newRepo(t)
		saveActor(t, repo, "000000000")
		a := saveActor(t, repo, "999999999")

		err := a.ChangeTaxID("000000000")
		require.NoError(t, err)
		got, err := repo.Update(t.Context(), a, 2)
		require.ErrorIs(t, err, actor.ErrVersionConflict)
		assert.Nil(t, got)

		want := actor.Reconstitute(a.ID(), "999999999", a.Type(), 1)
		checkActor(t, repo, want)
	})
}

func runDeleteContract(t *testing.T, newRepo func(t *testing.T) actor.Repository) {
	t.Helper()

	t.Run("Delete removes the actor", func(t *testing.T) {
		repo := newRepo(t)
		a := saveActor(t, repo, "000000000")
		other := saveActor(t, repo, "111111111")

		err := repo.Delete(t.Context(), a.ID(), 1)
		require.NoError(t, err)

		_, err = repo.FindByID(t.Context(), a.ID())
		require.ErrorIs(t, err, actor.ErrNotFound)
		checkActor(t, repo, other)
	})

	t.Run("Delete frees the tax ID", func(t *testing.T) {
		repo := newRepo(t)
		a := saveActor(t, repo, "000000000")

		err := repo.Delete(t.Context(), a.ID(), 1)
		require.NoError(t, err)
		saveActor(t, repo, "000000000")
	})

	t.Run("Delete accepts the version after an update", func(t *testing.T) {
		repo := newRepo(t)
		a := saveActor(t, repo, "000000000")
		_, err := repo.Update(t.Context(), a, 1)
		require.NoError(t, err)

		err = repo.Delete(t.Context(), a.ID(), 2)
		require.NoError(t, err)
	})

	t.Run("Delete fails if actor does not exist", func(t *testing.T) {
		repo := newRepo(t)

		err := repo.Delete(t.Context(), uuid.New(), 1)
		require.ErrorIs(t, err, actor.ErrNotFound)
	})

	t.Run("Delete fails if already deleted", func(t *testing.T) {
		repo := newRepo(t)
		a := saveActor(t, repo, "000000000")
		require.NoError(t, repo.Delete(t.Context(), a.ID(), 1))

		err := repo.Delete(t.Context(), a.ID(), 1)
		require.ErrorIs(t, err, actor.ErrNotFound)
	})

	t.Run("Delete fails if versions do not match", func(t *testing.T) {
		repo := newRepo(t)
		a := saveActor(t, repo, "000000000")

		err := repo.Delete(t.Context(), a.ID(), 2)
		require.ErrorIs(t, err, actor.ErrVersionConflict)
		checkActor(t, repo, a)
	})
}

func runFindContract(t *testing.T, newRepo func(t *testing.T) actor.Repository) {
	t.Helper()

	t.Run("FindByID returns actor if exists", func(t *testing.T) {
		repo := newRepo(t)
		a := saveActor(t, repo, "000000000")

		got, err := repo.FindByID(t.Context(), a.ID())
		require.NoError(t, err)
		assert.Equal(t, a, got)
	})

	t.Run("FindByID fails if actor not found", func(t *testing.T) {
		repo := newRepo(t)

		got, err := repo.FindByID(t.Context(), uuid.New())
		require.ErrorIs(t, err, actor.ErrNotFound)
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

		got, err := repo.FindAll(t.Context(), newPage(t, actor.MaxPageLimit, 0))
		require.NoError(t, err)
		assert.Equal(t, actors, got)
	})

	t.Run("FindAll paginates", func(t *testing.T) {
		repo := newRepo(t)
		a1 := saveActor(t, repo, "000000000")
		a2 := saveActor(t, repo, "111111111")
		a3 := saveActor(t, repo, "222222222")

		tests := []struct {
			name          string
			limit, offset int32
			want          []*actor.Actor
		}{
			{"set limit", 2, 0, []*actor.Actor{a1, a2}},
			{"limit bigger than number of actors", 10, 0, []*actor.Actor{a1, a2, a3}},
			{"set offset", actor.MaxPageLimit, 1, []*actor.Actor{a2, a3}},
			{"set limit and offset", 1, 1, []*actor.Actor{a2}},
			{"offset bigger than number of actors", actor.MaxPageLimit, 10, []*actor.Actor{}},
			{"max limit and offset", actor.MaxPageLimit, math.MaxInt32, []*actor.Actor{}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := repo.FindAll(t.Context(), newPage(t, tt.limit, tt.offset))
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			})
		}
	})
}

func runContextCanceledContract(t *testing.T, newRepo func(t *testing.T) actor.Repository) {
	t.Helper()

	t.Run("Context canceled", func(t *testing.T) {
		repo := newRepo(t)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		a := newActor(t, "000000000")
		page := newPage(t, actor.MaxPageLimit, 0)

		tests := []struct {
			name string
			call func() (any, error)
		}{
			{"Count", func() (any, error) { return repo.Count(ctx) }},
			{"Save", func() (any, error) { return nil, repo.Save(ctx, a) }},
			{"Update", func() (any, error) { return repo.Update(ctx, a, 1) }},
			{"Delete", func() (any, error) { return nil, repo.Delete(ctx, a.ID(), 1) }},
			{"FindByID", func() (any, error) { return repo.FindByID(ctx, a.ID()) }},
			{"FindAll", func() (any, error) { return repo.FindAll(ctx, page) }},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := tt.call()
				require.ErrorIs(t, err, context.Canceled)
				assert.Zero(t, got)
			})
		}
	})
}

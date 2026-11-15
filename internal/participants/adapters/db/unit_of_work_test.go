package db

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"exercise/internal/participants/app/uow"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
)

var errBoom = errors.New("boom")

func setupUoW(t *testing.T) (*UnitOfWork, account.Repository, uuid.UUID) {
	t.Helper()
	f := setup(t)
	return NewUnitOfWork(testPool), f.Repo, f.NewActorID(t)
}

func assertNotPersisted(t *testing.T, repo account.Repository, id uuid.UUID) {
	t.Helper()
	_, err := repo.FindByID(t.Context(), id)
	require.ErrorIs(t, err, account.ErrNotFound)
}

func TestUnitOfWork(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test")
	}

	t.Run("commits when fn succeeds", func(t *testing.T) {
		u, repo, actorID := setupUoW(t)
		acc := accounttest.NewAccount(t, actorID)

		err := u.Do(t.Context(), func(ctx context.Context, repos uow.Repositories) error {
			return repos.Accounts.Save(ctx, acc)
		})
		require.NoError(t, err)

		_, err = repo.FindByID(t.Context(), acc.ID())
		require.NoError(t, err)
	})

	t.Run("rolls back and returns fn error unchanged", func(t *testing.T) {
		u, repo, actorID := setupUoW(t)
		acc := accounttest.NewAccount(t, actorID)

		err := u.Do(t.Context(), func(ctx context.Context, repos uow.Repositories) error {
			err := repos.Accounts.Save(ctx, acc)
			require.NoError(t, err)
			return errBoom
		})
		require.ErrorIs(t, err, errBoom)

		assertNotPersisted(t, repo, acc.ID())
	})

	t.Run("rolls back on panic", func(t *testing.T) {
		u, repo, actorID := setupUoW(t)
		acc := accounttest.NewAccount(t, actorID)

		assert.PanicsWithValue(t, "boom", func() {
			err := u.Do(t.Context(), func(ctx context.Context, repos uow.Repositories) error {
				err := repos.Accounts.Save(ctx, acc)
				require.NoError(t, err)
				panic("boom")
			})
			require.NoError(t, err)
		})

		assertNotPersisted(t, repo, acc.ID())
	})

	t.Run("fails to begin with a canceled context", func(t *testing.T) {
		u, _, _ := setupUoW(t)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		called := false
		err := u.Do(ctx, func(context.Context, uow.Repositories) error {
			called = true
			return nil
		})
		require.ErrorIs(t, err, context.Canceled)
		require.False(t, called)
	})
}

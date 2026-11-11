package db

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"exercise/internal/participants/adapters/db/sqlcgen"
	"exercise/internal/participants/app/uow"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
)

var errBoom = errors.New("boom")

func setupUoW(t *testing.T) (*UnitOfWork, account.Repository, uuid.UUID) {
	t.Helper()
	_, _ = testPool.Exec(t.Context(), "TRUNCATE actors CASCADE")
	q := sqlcgen.New(testPool)
	return NewUnitOfWork(testPool), NewAccountRepository(q), seedActor(t, q)
}

func assertNotPersisted(t *testing.T, repo account.Repository, id uuid.UUID) {
	t.Helper()
	if _, err := repo.FindByID(t.Context(), id); !errors.Is(err, account.ErrAccountNotFound) {
		t.Errorf("FindByID() = %v, want %v", err, account.ErrAccountNotFound)
	}
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
		if err != nil {
			t.Fatalf("Do() = %v, want nil", err)
		}

		if _, err := repo.FindByID(t.Context(), acc.ID); err != nil {
			t.Errorf("FindByID() = %v, want nil", err)
		}
	})

	t.Run("rolls back and returns fn error unchanged", func(t *testing.T) {
		u, repo, actorID := setupUoW(t)
		acc := accounttest.NewAccount(t, actorID)

		err := u.Do(t.Context(), func(ctx context.Context, repos uow.Repositories) error {
			if err := repos.Accounts.Save(ctx, acc); err != nil {
				return err
			}
			return errBoom
		})
		if !errors.Is(err, errBoom) {
			t.Errorf("Do() = %v, want %v", err, errBoom)
		}

		assertNotPersisted(t, repo, acc.ID)
	})

	t.Run("rolls back on panic", func(t *testing.T) {
		u, repo, actorID := setupUoW(t)
		acc := accounttest.NewAccount(t, actorID)

		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Error("Do() did not propagate the panic")
				}
			}()
			_ = u.Do(t.Context(), func(ctx context.Context, repos uow.Repositories) error {
				if err := repos.Accounts.Save(ctx, acc); err != nil {
					return err
				}
				panic("boom")
			})
		}()

		assertNotPersisted(t, repo, acc.ID)
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
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Do() = %v, want %v", err, context.Canceled)
		}
		if called {
			t.Error("fn was called, want not called")
		}
	})
}

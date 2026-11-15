package accounttest

import (
	"context"
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"exercise/internal/participants/domain/account"
)

type Fixture struct {
	Repo       account.Repository
	NewActorID func(t *testing.T) uuid.UUID
}

func NewAccount(t *testing.T, actorID uuid.UUID) *account.Account {
	t.Helper()
	a, err := account.New(uuid.New(), actorID, "dltID", "privKey", "pubKey")
	require.NoError(t, err)
	return a
}

func saveAccount(t *testing.T, repo account.Repository, actorID uuid.UUID) *account.Account {
	t.Helper()
	a := NewAccount(t, actorID)
	err := repo.Save(t.Context(), a)
	require.NoError(t, err)
	return a
}

func checkAccount(t *testing.T, repo account.Repository, want *account.Account) {
	t.Helper()
	got, err := repo.FindByID(t.Context(), want.ID())
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func RunRepositoryContract(t *testing.T, setup func(t *testing.T) Fixture) {
	t.Helper()

	t.Run("Save stores the account", func(t *testing.T) {
		f := setup(t)
		a := NewAccount(t, f.NewActorID(t))

		err := f.Repo.Save(t.Context(), a)
		require.NoError(t, err)
		checkAccount(t, f.Repo, a)
	})

	t.Run("Save fails if ID already exists", func(t *testing.T) {
		f := setup(t)
		a := saveAccount(t, f.Repo, f.NewActorID(t))
		duplicate, err := account.New(a.ID(), a.ActorID(), "Other dltID", a.PrivateKey(), a.PublicKey())
		require.NoError(t, err)

		err = f.Repo.Save(t.Context(), duplicate)
		require.ErrorIs(t, err, account.ErrAlreadyExists)
		checkAccount(t, f.Repo, a)
	})

	t.Run("Save fails if context is canceled", func(t *testing.T) {
		f := setup(t)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		a := NewAccount(t, uuid.New())
		err := f.Repo.Save(ctx, a)
		require.ErrorIs(t, err, context.Canceled)
	})

	t.Run("FindByID fails if account not found", func(t *testing.T) {
		f := setup(t)

		got, err := f.Repo.FindByID(t.Context(), uuid.New())
		require.ErrorIs(t, err, account.ErrNotFound)
		assert.Nil(t, got)
	})

	t.Run("FindByID fails if context is canceled", func(t *testing.T) {
		f := setup(t)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		got, err := f.Repo.FindByID(ctx, uuid.New())
		require.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, got)
	})

	t.Run("FindAll returns all accounts in inserted order", func(t *testing.T) {
		f := setup(t)
		accounts := []*account.Account{
			saveAccount(t, f.Repo, f.NewActorID(t)),
			saveAccount(t, f.Repo, f.NewActorID(t)),
			saveAccount(t, f.Repo, f.NewActorID(t)),
			saveAccount(t, f.Repo, f.NewActorID(t)),
		}

		got, err := f.Repo.FindAll(t.Context())
		require.NoError(t, err)
		assert.Equal(t, accounts, got)
	})

	t.Run("FindAll fails if context is canceled", func(t *testing.T) {
		f := setup(t)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		got, err := f.Repo.FindAll(ctx)
		require.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, got)
	})

	t.Run("FindByActorIDs returns accounts filtered by actor ID", func(t *testing.T) {
		f := setup(t)
		actorID1 := f.NewActorID(t)
		actorID2 := f.NewActorID(t)
		accounts := []*account.Account{
			saveAccount(t, f.Repo, actorID1),
			saveAccount(t, f.Repo, actorID1),
			saveAccount(t, f.Repo, actorID2),
		}
		saveAccount(t, f.Repo, f.NewActorID(t))

		got, err := f.Repo.FindByActorIDs(t.Context(), []uuid.UUID{actorID1, actorID2})
		require.NoError(t, err)
		assert.Equal(t, accounts, got)
	})

	t.Run("FindByActorIDs returns empty if no actor IDs", func(t *testing.T) {
		f := setup(t)

		got, err := f.Repo.FindByActorIDs(t.Context(), []uuid.UUID{})
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("FindByActorIDs fails if context is canceled", func(t *testing.T) {
		f := setup(t)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		got, err := f.Repo.FindByActorIDs(ctx, []uuid.UUID{uuid.New()})
		require.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, got)
	})
}

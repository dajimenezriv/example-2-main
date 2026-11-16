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

	t.Run("FindByID fails if account not found", func(t *testing.T) {
		f := setup(t)

		got, err := f.Repo.FindByID(t.Context(), uuid.New())
		require.ErrorIs(t, err, account.ErrNotFound)
		assert.Nil(t, got)
	})

	t.Run("FindAll errors", func(t *testing.T) {
		f := setup(t)

		tests := []struct {
			name          string
			limit, offset int
		}{
			{"negative limit", -1, 0},
			{"negative offset", 0, -1},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := f.Repo.FindAll(t.Context(), tt.limit, tt.offset)
				require.ErrorIs(t, err, account.ErrInvalidPagination)
				assert.Nil(t, got)
			})
		}
	})

	t.Run("FindAll success", func(t *testing.T) {
		f := setup(t)
		a1 := saveAccount(t, f.Repo, f.NewActorID(t))
		a2 := saveAccount(t, f.Repo, f.NewActorID(t))
		a3 := saveAccount(t, f.Repo, f.NewActorID(t))

		tests := []struct {
			name          string
			limit, offset int
			want          []*account.Account
		}{
			{"zero limit", 0, 0, []*account.Account{}},
			{"set limit", 2, 0, []*account.Account{a1, a2}},
			{"limit bigger than number of accounts", 10, 0, []*account.Account{a1, a2, a3}},
			{"set offset", 100, 1, []*account.Account{a2, a3}},
			{"offset bigger than number of accounts", 100, 10, []*account.Account{}},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := f.Repo.FindAll(t.Context(), tt.limit, tt.offset)
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			})
		}
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

	t.Run("Context canceled", func(t *testing.T) {
		f := setup(t)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		tests := []struct {
			name string
			fn   func() (any, error)
		}{
			{
				name: "Save",
				fn: func() (any, error) {
					a := NewAccount(t, uuid.New())
					err := f.Repo.Save(ctx, a)
					return nil, err
				},
			},
			{
				name: "FindByID",
				fn:   func() (any, error) { return f.Repo.FindByID(ctx, uuid.New()) },
			},
			{
				name: "FindAll",
				fn:   func() (any, error) { return f.Repo.FindAll(ctx, 100, 0) },
			},
			{
				name: "FindByActorIDs",
				fn:   func() (any, error) { return f.Repo.FindByActorIDs(ctx, []uuid.UUID{uuid.New()}) },
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := tt.fn()
				require.ErrorIs(t, err, context.Canceled)
				assert.Nil(t, got)
			})
		}
	})
}

package accounttest

import (
	"context"
	"errors"
	"exercise/internal/participants/domain/account"
	"testing"
	"uuid"

	"github.com/google/go-cmp/cmp"
)

type Fixture struct {
	Repo       account.Repository
	NewActorID func(t *testing.T) uuid.UUID
}

func NewAccount(t *testing.T, actorID uuid.UUID) account.Account {
	t.Helper()
	a, _ := account.New(uuid.New(), actorID, "DLT account ID", "Private key", "Public key")
	return a
}

func saveAccount(t *testing.T, repo account.Repository, actorID uuid.UUID) *account.Account {
	t.Helper()
	a := NewAccount(t, actorID)
	_ = repo.Save(t.Context(), &a)
	return &a
}

func RunRepositoryContract(t *testing.T, setup func(t *testing.T) Fixture) {
	t.Run("Save stores the account", func(t *testing.T) {
		f := setup(t)
		want := NewAccount(t, f.NewActorID(t))

		if err := f.Repo.Save(t.Context(), &want); err != nil {
			t.Fatalf("Save() = %v, want nil", err)
		}

		got, err := f.Repo.FindByID(t.Context(), want.ID)
		if err != nil {
			t.Errorf("FindByID() = %v, want nil", err)
		}
		if diff := cmp.Diff(&want, got); diff != "" {
			t.Errorf("account (-want +got):\n%s", diff)
		}
	})

	t.Run("Save fails if account already exists", func(t *testing.T) {
		f := setup(t)
		want := NewAccount(t, f.NewActorID(t))

		if err := f.Repo.Save(t.Context(), &want); err != nil {
			t.Fatalf("Save() = %v, want nil", err)
		}

		duplicate := want
		duplicate.DltAccountID = "Other DLT account ID"
		if err := f.Repo.Save(t.Context(), &duplicate); !errors.Is(err, account.ErrAccountAlreadyExists) {
			t.Fatalf("Save() = %v, want %v", err, account.ErrAccountAlreadyExists)
		}

		got, err := f.Repo.FindByID(t.Context(), want.ID)
		if err != nil {
			t.Errorf("FindByID() = %v, want nil", err)
		}
		if diff := cmp.Diff(&want, got); diff != "" {
			t.Errorf("account (-want +got):\n%s", diff)
		}
	})

	t.Run("Save returns error if context is canceled", func(t *testing.T) {
		f := setup(t)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		err := f.Repo.Save(ctx, new(NewAccount(t, uuid.New())))

		if !errors.Is(err, context.Canceled) {
			t.Errorf("FindByID() = %v, want %v", err, context.Canceled)
		}
	})

	t.Run("FindByID returns error if account not found", func(t *testing.T) {
		f := setup(t)
		a, err := f.Repo.FindByID(t.Context(), uuid.New())

		if !errors.Is(err, account.ErrAccountNotFound) {
			t.Errorf("FindByID() = %v, want %v", err, account.ErrAccountNotFound)
		}
		if a != nil {
			t.Errorf("FindByID() = %+v, want nil", a)
		}
	})

	t.Run("FindByID returns error if context is canceled", func(t *testing.T) {
		f := setup(t)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		got, err := f.Repo.FindByID(ctx, uuid.New())

		if !errors.Is(err, context.Canceled) {
			t.Errorf("FindByID() = %v, want %v", err, context.Canceled)
		}
		if got != nil {
			t.Errorf("FindByID() = %+v, want nil", got)
		}
	})

	t.Run("FindAll returns all accounts in inserted order", func(t *testing.T) {
		f := setup(t)
		want := make([]*account.Account, 0, 4)
		want = append(want, saveAccount(t, f.Repo, f.NewActorID(t)))
		want = append(want, saveAccount(t, f.Repo, f.NewActorID(t)))
		want = append(want, saveAccount(t, f.Repo, f.NewActorID(t)))
		want = append(want, saveAccount(t, f.Repo, f.NewActorID(t)))

		got, err := f.Repo.FindAll(t.Context())
		if err != nil {
			t.Errorf("FindAll() = %v, want nil", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("accounts (-want +got):\n%s", diff)
		}
	})

	t.Run("FindAll returns error if context is canceled", func(t *testing.T) {
		f := setup(t)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		got, err := f.Repo.FindAll(ctx)

		if !errors.Is(err, context.Canceled) {
			t.Errorf("FindAll() = %v, want %v", err, context.Canceled)
		}
		if got != nil {
			t.Errorf("FindAll() = %+v, want nil", got)
		}
	})

	t.Run("FindByActorIDs returns accounts filtered by actor ID", func(t *testing.T) {
		f := setup(t)
		actorID1 := f.NewActorID(t)
		actorID2 := f.NewActorID(t)
		want := make([]*account.Account, 0, 4)
		want = append(want, saveAccount(t, f.Repo, actorID1))
		want = append(want, saveAccount(t, f.Repo, actorID1))
		want = append(want, saveAccount(t, f.Repo, actorID2))
		saveAccount(t, f.Repo, f.NewActorID(t))

		got, err := f.Repo.FindByActorIDs(t.Context(), []uuid.UUID{actorID1, actorID2})
		if err != nil {
			t.Errorf("FindAll() = %v, want nil", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("accounts (-want +got):\n%s", diff)
		}
	})

	t.Run("FindByActorIDs returns empty if no actor uuids", func(t *testing.T) {
		f := setup(t)

		got, err := f.Repo.FindByActorIDs(t.Context(), []uuid.UUID{})

		if err != nil {
			t.Errorf("FindByActorIDs() = %v, want nil", err)
		}
		if got := len(got); got != 0 {
			t.Errorf("len(FindByActorIDs()) = %d, want 0", got)
		}
	})

	t.Run("FindByActorIDs returns error if context is canceled", func(t *testing.T) {
		f := setup(t)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		got, err := f.Repo.FindByActorIDs(ctx, []uuid.UUID{uuid.New()})

		if !errors.Is(err, context.Canceled) {
			t.Errorf("FindByActorIDs() = %v, want %v", err, context.Canceled)
		}
		if got != nil {
			t.Errorf("FindByActorIDs() = %+v, want nil", got)
		}
	})
}

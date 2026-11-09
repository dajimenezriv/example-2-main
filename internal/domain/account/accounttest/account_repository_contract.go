package accounttest

import (
	"errors"
	"exercise/internal/domain/account"
	"testing"
	"uuid"

	"github.com/google/go-cmp/cmp"
)

type Fixture struct {
	Repo       account.Repository
	NewActorID func(t *testing.T) uuid.UUID
}

func newAccount(t *testing.T, actorID uuid.UUID) account.Account {
	t.Helper()
	return account.New(uuid.New(), actorID, "DLT account ID", "Private key", "Public key")
}

func saveAccount(t *testing.T, repo account.Repository, actorID uuid.UUID) *account.Account {
	t.Helper()
	a := newAccount(t, actorID)
	if err := repo.Save(&a); err != nil {
		t.Fatalf("repo.Save = %v", err)
	}
	return &a
}

func RunRepositoryContract(t *testing.T, setup func(t *testing.T) Fixture) {
	t.Run("Save stores the account", func(t *testing.T) {
		f := setup(t)
		want := newAccount(t, f.NewActorID(t))

		if err := f.Repo.Save(&want); err != nil {
			t.Fatalf("Save() = %v, want nil", err)
		}

		got, err := f.Repo.FindByID(want.ID)
		if err != nil {
			t.Errorf("FindByID() = %v, want nil", err)
		}
		if diff := cmp.Diff(&want, got); diff != "" {
			t.Errorf("account (-want +got):\n%s", diff)
		}
	})

	t.Run("Save fails if account already exists", func(t *testing.T) {
		f := setup(t)
		want := newAccount(t, f.NewActorID(t))

		if err := f.Repo.Save(&want); err != nil {
			t.Fatalf("Save() = %v, want nil", err)
		}

		duplicate := want
		duplicate.DltAccountId = "Other DLT account ID"
		if err := f.Repo.Save(&duplicate); !errors.Is(err, account.ErrAccountAlreadyExists) {
			t.Fatalf("Save() = %v, want %v", err, account.ErrAccountAlreadyExists)
		}

		got, err := f.Repo.FindByID(want.ID)
		if err != nil {
			t.Errorf("FindByID() = %v, want nil", err)
		}
		if diff := cmp.Diff(&want, got); diff != "" {
			t.Errorf("account (-want +got):\n%s", diff)
		}
	})

	t.Run("FindByID returns error if account not found", func(t *testing.T) {
		f := setup(t)
		a, err := f.Repo.FindByID(uuid.New())

		if !errors.Is(err, account.ErrAccountNotFound) {
			t.Errorf("FindByID() = %v, want %v", err, account.ErrAccountNotFound)
		}
		if a != nil {
			t.Errorf("FindByID() = %+v, want nil", a)
		}
	})

	t.Run("FindAll returns all accounts in inserted order", func(t *testing.T) {
		f := setup(t)
		want := make([]*account.Account, 0, 4)
		want = append(want, saveAccount(t, f.Repo, f.NewActorID(t)))
		want = append(want, saveAccount(t, f.Repo, f.NewActorID(t)))
		want = append(want, saveAccount(t, f.Repo, f.NewActorID(t)))
		want = append(want, saveAccount(t, f.Repo, f.NewActorID(t)))

		got, err := f.Repo.FindAll()
		if err != nil {
			t.Errorf("FindAll() = %v, want nil", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("accounts (-want +got):\n%s", diff)
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

		got, err := f.Repo.FindByActorIDs([]uuid.UUID{actorID1, actorID2})
		if err != nil {
			t.Errorf("FindAll() = %v, want nil", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("accounts (-want +got):\n%s", diff)
		}
	})
}

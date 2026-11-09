package accounttest

import (
	"errors"
	"exercise2/cmd/actormgmt/internal/domain/account"
	"testing"
	"uuid"

	"github.com/google/go-cmp/cmp"
)

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

func RunRepositoryContract(t *testing.T, newRepo func(t *testing.T) account.Repository) {
	t.Run("Save stores the account", func(t *testing.T) {
		repo := newRepo(t)
		want := newAccount(t, uuid.New())

		if err := repo.Save(&want); err != nil {
			t.Fatalf("Save() = %v, want nil", err)
		}

		got, err := repo.FindByID(want.ID)
		if err != nil {
			t.Errorf("FindByID() = %v, want nil", err)
		}
		if diff := cmp.Diff(&want, got); diff != "" {
			t.Errorf("account (-want +got):\n%s", diff)
		}
	})

	t.Run("Save fails if account already exists", func(t *testing.T) {
		repo := newRepo(t)
		want := newAccount(t, uuid.New())

		if err := repo.Save(&want); err != nil {
			t.Fatalf("Save() = %v, want nil", err)
		}

		duplicate := want
		duplicate.DltAccountId = "Other DLT account ID"
		if err := repo.Save(&duplicate); !errors.Is(err, account.ErrAccountAlreadyExists) {
			t.Fatalf("Save() = %v, want %v", err, account.ErrAccountAlreadyExists)
		}

		got, err := repo.FindByID(want.ID)
		if err != nil {
			t.Errorf("FindByID() = %v, want nil", err)
		}
		if diff := cmp.Diff(&want, got); diff != "" {
			t.Errorf("account (-want +got):\n%s", diff)
		}
	})

	t.Run("FindByID returns error if account not found", func(t *testing.T) {
		repo := newRepo(t)
		a, err := repo.FindByID(uuid.New())

		if !errors.Is(err, account.ErrAccountNotFound) {
			t.Errorf("FindByID() = %v, want %v", err, account.ErrAccountNotFound)
		}
		if a != nil {
			t.Errorf("FindByID() = %+v, want nil", a)
		}
	})

	t.Run("FindAll returns all accounts in inserted order", func(t *testing.T) {
		repo := newRepo(t)
		want := make([]*account.Account, 0, 4)
		want = append(want, saveAccount(t, repo, uuid.New()))
		want = append(want, saveAccount(t, repo, uuid.New()))
		want = append(want, saveAccount(t, repo, uuid.New()))
		want = append(want, saveAccount(t, repo, uuid.New()))

		got, err := repo.FindAll()
		if err != nil {
			t.Errorf("FindAll() = %v, want nil", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("accounts (-want +got):\n%s", diff)
		}
	})

	t.Run("FindByActorIDs returns accounts filtered by actor ID", func(t *testing.T) {
		actorID1 := uuid.New()
		actorID2 := uuid.New()
		repo := newRepo(t)
		want := make([]*account.Account, 0, 4)
		want = append(want, saveAccount(t, repo, actorID1))
		want = append(want, saveAccount(t, repo, actorID1))
		want = append(want, saveAccount(t, repo, actorID2))
		saveAccount(t, repo, uuid.New())

		got, err := repo.FindByActorIDs([]uuid.UUID{actorID1, actorID2})
		if err != nil {
			t.Errorf("FindAll() = %v, want nil", err)
		}
		if diff := cmp.Diff(want, got); diff != "" {
			t.Errorf("accounts (-want +got):\n%s", diff)
		}
	})
}

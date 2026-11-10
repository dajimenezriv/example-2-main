package db

import (
	"errors"
	"exercise/internal/participants/adapters/db/sqlcgen"
	"exercise/internal/participants/domain/account/accounttest"
	"testing"
	"uuid"

	"github.com/jackc/pgx/v5/pgconn"
)

func seedActor(t *testing.T, q *sqlcgen.Queries) uuid.UUID {
	t.Helper()
	id := uuid.New()
	_ = q.CreateActor(t.Context(), sqlcgen.CreateActorParams{ID: id, TaxID: "Tax ID", Type: "Type"})
	return id
}

func setup(t *testing.T) accounttest.Fixture {
	t.Helper()
	q := sqlcgen.New(testPool)
	// Truncating actors also empties accounts through CASCADE.
	_, _ = testPool.Exec(t.Context(), "TRUNCATE actors CASCADE")
	return accounttest.Fixture{
		Repo: NewAccountRepository(q),
		NewActorID: func(t *testing.T) uuid.UUID {
			t.Helper()
			return seedActor(t, q)
		},
	}
}

func TestAccountRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test")
	}

	accounttest.RunRepositoryContract(t, setup)

	t.Run("Save fails if actor does not exist", func(t *testing.T) {
		f := setup(t)
		a := accounttest.NewAccount(t, uuid.New())

		err := f.Repo.Save(t.Context(), &a)
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); !ok || pgErr.Code != "23503" {
			t.Errorf("Save() = %v, want FK violation error", err)
		}
	})
}

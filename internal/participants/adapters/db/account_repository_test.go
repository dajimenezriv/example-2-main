package db

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"uuid"

	"exercise/internal/participants/adapters/db/sqlcgen"
	"exercise/internal/participants/domain/account/accounttest"

	"github.com/jackc/pgx/v5/pgconn"
)

var seedTaxID atomic.Int64

func seedActor(t *testing.T, q *sqlcgen.Queries) uuid.UUID {
	t.Helper()
	id := uuid.New()
	taxID := fmt.Sprintf("%09d", seedTaxID.Add(1))
	if err := q.CreateActor(t.Context(), sqlcgen.CreateActorParams{ID: id, TaxID: taxID, Type: "issuer"}); err != nil {
		t.Fatalf("seed actor: %v", err)
	}
	return id
}

func setup(t *testing.T) accounttest.Fixture {
	t.Helper()
	q := sqlcgen.New(testPool)
	// Truncating actors also empties accounts.
	if _, err := testPool.Exec(t.Context(), "TRUNCATE actors CASCADE"); err != nil {
		t.Fatalf("truncate actors: %v", err)
	}
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

		err := f.Repo.Save(t.Context(), a)
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); !ok || pgErr.Code != foreignKeyViolation {
			t.Errorf("Save() = %v, want FK violation error", err)
		}
	})
}

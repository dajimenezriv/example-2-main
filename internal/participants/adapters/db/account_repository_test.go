package db

import (
	"exercise/internal/participants/adapters/db/sqlcgen"
	"exercise/internal/participants/domain/account/accounttest"
	"testing"
	"uuid"
)

func seedActor(t *testing.T, q *sqlcgen.Queries) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if err := q.CreateActor(t.Context(), sqlcgen.CreateActorParams{ID: id, TaxID: "Tax ID", Type: "Type"}); err != nil {
		t.Fatalf("seed actor: %v", err)
	}
	return id
}

func TestAccountRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test")
	}

	q := sqlcgen.New(testPool)

	accounttest.RunRepositoryContract(t, func(t *testing.T) accounttest.Fixture {
		t.Helper()
		// Truncating actors also empties accounts through CASCADE.
		if _, err := testPool.Exec(t.Context(), "TRUNCATE actors CASCADE"); err != nil {
			t.Fatalf("TRUNCATE actors: %v", err)
		}
		return accounttest.Fixture{
			Repo: NewAccountRepository(q),
			NewActorID: func(t *testing.T) uuid.UUID {
				t.Helper()
				return seedActor(t, q)
			},
		}
	})
}

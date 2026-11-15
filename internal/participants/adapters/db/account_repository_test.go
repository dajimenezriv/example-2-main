package db

import (
	"fmt"
	"sync/atomic"
	"testing"
	"uuid"

	"exercise/internal/participants/adapters/db/sqlcgen"
	"exercise/internal/participants/domain/account/accounttest"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var seedTaxID atomic.Int64

func seedActor(t *testing.T, q *sqlcgen.Queries) uuid.UUID {
	t.Helper()
	id := uuid.New()
	taxID := fmt.Sprintf("%09d", seedTaxID.Add(1))
	err := q.CreateActor(t.Context(), sqlcgen.CreateActorParams{ID: id, TaxID: taxID, Type: "issuer"})
	require.NoError(t, err, "seed actor")
	return id
}

func setup(t *testing.T) accounttest.Fixture {
	t.Helper()
	q := sqlcgen.New(testPool)
	// Truncating actors also empties accounts.
	_, err := testPool.Exec(t.Context(), "TRUNCATE actors CASCADE")
	require.NoError(t, err, "truncate actors")
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
		var pgErr *pgconn.PgError
		require.ErrorAs(t, err, &pgErr)
		assert.Equal(t, foreignKeyViolation, pgErr.Code)
	})
}

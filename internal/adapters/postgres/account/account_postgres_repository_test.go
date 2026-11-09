package accountrepo

import (
	"context"
	"exercise/internal/adapters/postgres/sqlcgen"
	"exercise/internal/domain/account/accounttest"
	"log"
	"testing"
	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func seedActor(t *testing.T, q *sqlcgen.Queries) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if err := q.CreateActor(t.Context(), sqlcgen.CreateActorParams{ID: id, TaxID: "Tax ID", Type: "Type"}); err != nil {
		t.Fatalf("seed actor: %v", err)
	}
	return id
}

func TestPostgresRepositoryContract(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test")
	}

	ctx := context.Background()

	ctr, err := postgres.Run(ctx, "postgres:16",
		postgres.WithInitScripts("../../../../docker/postgres/init.sql"),
		postgres.BasicWaitStrategies())
	testcontainers.CleanupContainer(t, ctr)
	if err != nil {
		t.Fatalf("postgres.Run = %v", err)
	}

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("ctr.ConnectionString = %v", err)
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("pgxpool.New: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	q := sqlcgen.New(pool)

	accounttest.RunRepositoryContract(t, func(t *testing.T) accounttest.Fixture {
		t.Helper()
		// Truncating actors also empties accounts through CASCADE.
		if _, err := pool.Exec(t.Context(), "TRUNCATE actors CASCADE"); err != nil {
			t.Fatalf("TRUNCATE actors: %v", err)
		}
		return accounttest.Fixture{
			Repo: NewAccountPostgresRepository(q),
			NewActorID: func(t *testing.T) uuid.UUID {
				t.Helper()
				return seedActor(t, q)
			},
		}
	})
}

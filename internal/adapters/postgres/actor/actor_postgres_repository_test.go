package actorrepo

import (
	"context"
	"exercise/internal/adapters/postgres/sqlcgen"
	"exercise/internal/domain/actor"
	"exercise/internal/domain/actor/actortest"
	"log"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

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

	actortest.RunRepositoryContract(t, func(t *testing.T) actor.Repository {
		t.Helper()
		if _, err := pool.Exec(t.Context(), "TRUNCATE actors CASCADE"); err != nil {
			t.Fatalf("TRUNCATE actors: %v", err)
		}
		return NewPostgresRepository(q)
	})
}

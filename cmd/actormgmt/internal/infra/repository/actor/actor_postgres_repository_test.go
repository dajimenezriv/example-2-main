package actorrepo

import (
	"context"
	"database/sql"
	"exercise2/cmd/actormgmt/internal/domain/actor"
	"exercise2/cmd/actormgmt/internal/domain/actor/actortest"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestPostgresRepositoryContract(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test")
	}

	ctx := context.Background()

	ctr, err := postgres.Run(ctx, "postgres:16",
		postgres.WithInitScripts("../../../../../../docker/postgres/init.sql"),
		postgres.BasicWaitStrategies())
	testcontainers.CleanupContainer(t, ctr)
	if err != nil {
		t.Fatalf("postgres.Run = %v", err)
	}

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("ctr.ConnectionString = %v", err)
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open = %v", err)
	}
	t.Cleanup(func() { db.Close() })

	actortest.RunRepositoryContract(t, func(t *testing.T) actor.Repository {
		t.Helper()
		if _, err := db.Exec("TRUNCATE actors CASCADE"); err != nil {
			t.Fatalf("TRUNCATE actors: %v", err)
		}
		return NewPostgresRepository(db)
	})
}

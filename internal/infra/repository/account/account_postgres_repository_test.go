package accountrepo

import (
	"context"
	"database/sql"
	"exercise2/internal/domain/account/accounttest"
	"testing"
	"uuid"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func seedActor(t *testing.T, db *sql.DB) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := db.Exec("INSERT INTO actors (id, tax_id, type) VALUES ($1, 'Tax ID', 'Type')", id); err != nil {
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

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open = %v", err)
	}
	t.Cleanup(func() { db.Close() })

	accounttest.RunRepositoryContract(t, func(t *testing.T) accounttest.Fixture {
		t.Helper()
		// Truncating actors also empties accounts through CASCADE.
		if _, err := db.Exec("TRUNCATE actors CASCADE"); err != nil {
			t.Fatalf("TRUNCATE actors: %v", err)
		}
		return accounttest.Fixture{
			Repo: NewPostgresRepository(db),
			NewActorID: func(t *testing.T) uuid.UUID {
				t.Helper()
				return seedActor(t, db)
			},
		}
	})
}

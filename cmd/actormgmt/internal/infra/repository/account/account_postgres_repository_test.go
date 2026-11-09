package accountrepo

import (
	"context"
	"database/sql"
	"exercise2/cmd/actormgmt/internal/domain/account"
	"exercise2/cmd/actormgmt/internal/domain/account/accounttest"
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

	accounttest.RunRepositoryContract(t, func(t *testing.T) account.Repository {
		t.Helper()
		if _, err := db.Exec("TRUNCATE accounts CASCADE"); err != nil {
			t.Fatalf("TRUNCATE accounts: %v", err)
		}
		return NewPostgresRepository(db)
	})
}

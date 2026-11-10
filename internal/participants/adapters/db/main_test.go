package db

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const migrationsGlob = "../../../../migrations/*.up.sql"

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	flag.Parse()
	if testing.Short() {
		return m.Run()
	}

	ctx := context.Background()

	migrations, err := filepath.Glob(migrationsGlob)
	if err != nil {
		log.Printf("filepath.Glob: %v", err)
		return 1
	}

	ctr, err := postgres.Run(ctx, "postgres:16",
		postgres.WithInitScripts(migrations...),
		postgres.BasicWaitStrategies())
	defer func() {
		if err := testcontainers.TerminateContainer(ctr); err != nil {
			log.Printf("terminate postgres container: %v", err)
		}
	}()
	if err != nil {
		log.Printf("start postgres container: %v", err)
		return 1
	}

	dsn, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Printf("build connection string: %v", err)
		return 1
	}
	testPool, err = pgxpool.New(ctx, dsn)
	if err != nil {
		log.Printf("create postgres pool: %v", err)
		return 1
	}
	defer testPool.Close()

	return m.Run()
}

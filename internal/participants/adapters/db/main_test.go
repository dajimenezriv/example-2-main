package db

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"testing"

	"exercise/migrations"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

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

	ctr, err := postgres.Run(ctx, "postgres:16", postgres.BasicWaitStrategies())
	defer func() {
		if termErr := testcontainers.TerminateContainer(ctr); termErr != nil {
			log.Printf("terminate postgres container: %v", termErr)
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

	// Same runner as `server migrate` in the deploy pipeline.
	if err := migrate(ctx); err != nil {
		log.Printf("migrate test DB: %v", err)
		return 1
	}

	return m.Run()
}

func migrate(ctx context.Context) (err error) {
	// Closing this *sql.DB leaves testPool open.
	db := stdlib.OpenDBFromPool(testPool)
	defer func() { err = errors.Join(err, db.Close()) }()

	p, err := migrations.NewProvider(db)
	if err != nil {
		return err
	}
	_, err = p.Up(ctx)
	return err
}

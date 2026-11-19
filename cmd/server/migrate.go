package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"exercise/migrations"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
)

func migrateUp() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	connCfg, err := pgx.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("parse DATABASE_URL: %w", err)
	}
	db := stdlib.OpenDB(*connCfg)
	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			log.Printf("close DB: %v", closeErr)
		}
	}()

	p, err := migrations.NewProvider(db)
	if err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	results, err := p.Up(ctx)
	for _, r := range results {
		log.Println(r)
	}
	if err != nil {
		return fmt.Errorf("migrate up: %w", err)
	}
	v, err := p.GetDBVersion(ctx)
	if err != nil {
		return fmt.Errorf("get DB version: %w", err)
	}
	log.Printf("✅ At version %d", v)
	return nil
}

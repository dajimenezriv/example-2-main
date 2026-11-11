package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"exercise/internal/platform/httpapi"

	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 120 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Printf("❌ %v", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	connStr := "postgres://ddd:ddd@localhost:5555/ddd?sslmode=disable"
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return fmt.Errorf("create DB pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping DB: %w", err)
	}
	log.Println("✅ Connected to Postgres")

	app := setupApplication(pool)

	srv := &http.Server{
		Addr:              ":8080",
		Handler:           registerHTTPHandlers(app),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	serveErr := make(chan error, 1)
	go func() {
		log.Println("🚀 Server running on :8080")
		serveErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen server: %w", err)
		}
		return nil
	case <-ctx.Done():
		log.Println("🛑 Shutting down")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}
	return nil
}

func registerHTTPHandlers(app *application) chi.Router {
	r := chi.NewRouter()

	api := humachi.New(r, httpapi.Config())

	app.GetActorsHTTP.Register(api)
	app.CreateActorHTTP.Register(api)
	app.GetAccountsHTTP.Register(api)
	app.CreateAccountHTTP.Register(api)

	return r
}

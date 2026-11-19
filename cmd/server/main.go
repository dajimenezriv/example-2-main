package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"exercise/internal/platform/httpapi"

	"github.com/caarlos0/env/v11"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 120 * time.Second
	shutdownTimeout   = 10 * time.Second
)

type config struct {
	DatabaseURL string `env:"DATABASE_URL,required,notEmpty"`
}

func main() {
	if err := dispatch(os.Args[1:]); err != nil {
		log.Printf("❌ %v", err)
		os.Exit(1)
	}
}

func dispatch(args []string) error {
	if len(args) == 0 {
		return run()
	}
	switch args[0] {
	case "openapi": // print the spec
		return printOpenAPI(os.Stdout)
	case "migrate": // apply pending migrations
		return migrateUp()
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func loadConfig() (config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return config{}, fmt.Errorf("load .env: %w", err)
	}
	cfg, err := env.ParseAs[config]()
	if err != nil {
		return config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
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
	registerOperations(humachi.New(r, httpapi.Config()), app)
	return r
}

// registerOperations is shared by the server and `openapi` (spec export), so the spec
// always lists exactly the endpoints the server serves.
func registerOperations(api huma.API, app *application) {
	app.GetActorsHTTP.Register(api)
	app.CreateActorHTTP.Register(api)
	app.UpdateActorHTTP.Register(api)
	app.DeleteActorHTTP.Register(api)
	app.GetAccountsHTTP.Register(api)
	app.CreateAccountHTTP.Register(api)
}

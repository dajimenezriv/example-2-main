package main

import (
	"context"
	"log"
	"net/http"

	"exercise/internal/platform/httpapi"

	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	connStr := "postgres://ddd:ddd@localhost:5432/ddd?sslmode=disable"
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("failed to create DB pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping DB: %v", err)
	}
	log.Println("✅ Connected to Postgres")

	app, err := setupApplication(pool)
	if err != nil {
		log.Fatalf("failed to setup context: %v", err)
	}

	router := registerHTTPHandlers(app)

	log.Println("🚀 Server running on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("server failed: %v", err)
	}
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

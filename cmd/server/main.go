package main

import (
	"context"
	"log"
	"net/http"

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

	mux := registerHTTPHandlers(app)

	log.Println("🚀 Server running on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func registerHTTPHandlers(app *application) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /actors", app.GetActorsHTTP.ServeHTTP)
	mux.HandleFunc("POST /actors", app.CreateActorHTTP.ServeHTTP)
	mux.HandleFunc("GET /accounts", app.GetAccountsHTTP.ServeHTTP)
	mux.HandleFunc("POST /accounts", app.CreateAccountHTTP.ServeHTTP)
	return mux
}

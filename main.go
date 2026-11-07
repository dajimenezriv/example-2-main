package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"exercise2/cmd/actormgmt"
)

func main() {
	connStr := "host=localhost port=5432 user=ddd password=ddd dbname=ddd sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping DB: %v", err)
	}
	log.Println("✅ Connected to Postgres")

	ctx, err := actormgmt.SetupContext(db)
	if err != nil {
		log.Fatalf("failed to setup context: %v", err)
	}

	mux := registerHTTPHandlers(ctx)

	log.Println("🚀 Server running on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func registerHTTPHandlers(ctx *actormgmt.Context) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /actors", ctx.GetActorsHTTP.ServeHTTP)
	mux.HandleFunc("POST /actors", ctx.CreateActorHTTP.ServeHTTP)
	mux.HandleFunc("GET /accounts", ctx.GetAccountsHTTP.ServeHTTP)
	mux.HandleFunc("POST /accounts", ctx.CreateAccountHTTP.ServeHTTP)
	return mux
}

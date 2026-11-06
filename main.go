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

	registerHTTPHandlers(ctx)

	log.Println("🚀 Server running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func registerHTTPHandlers(ctx *actormgmt.Context) {
	http.HandleFunc("/actors", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			ctx.CreateActorHTTP.ServeHTTP(w, r)
		case http.MethodGet:
			ctx.GetActorsHTTP.ServeHTTP(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/accounts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			ctx.CreateAccountHTTP.ServeHTTP(w, r)
		case http.MethodGet:
			ctx.GetAccountsHTTP.ServeHTTP(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

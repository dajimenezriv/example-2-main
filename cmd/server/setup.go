package main

import (
	"exercise/internal/adapters/db"
	"exercise/internal/adapters/db/sqlcgen"
	createAccountEndpoint "exercise/internal/adapters/http/createaccount"
	createActorEndpoint "exercise/internal/adapters/http/createactor"
	getAccountsEndpoint "exercise/internal/adapters/http/getaccounts"
	getActorsEndpoint "exercise/internal/adapters/http/getactors"
	createAccountCommand "exercise/internal/app/command/createaccount"
	createActorCommand "exercise/internal/app/command/createactor"
	"exercise/internal/app/query/getaccounts"
	"exercise/internal/app/query/getactors"

	"uuid"

	"github.com/jackc/pgx/v5/pgxpool"
)

type application struct {
	CreateActorHandler   *createActorCommand.Handler
	CreateAccountHandler *createAccountCommand.Handler
	GetAccountsQuery     *getaccounts.Query
	GetActorsQuery       *getactors.Query
	CreateAccountHTTP    *createAccountEndpoint.Handler
	CreateActorHTTP      *createActorEndpoint.Handler
	GetAccountsHTTP      *getAccountsEndpoint.Handler
	GetActorsHTTP        *getActorsEndpoint.Handler
}

func setupApplication(pool *pgxpool.Pool) (*application, error) {
	q := sqlcgen.New(pool)
	accountRepo := db.NewAccountRepository(q)
	actorRepo := db.NewActorRepository(q)

	newID := func() uuid.UUID { return uuid.NewV7() }
	createAccountHandler := createAccountCommand.NewHandler(accountRepo, actorRepo, newID)
	createActorHandler := createActorCommand.NewHandler(actorRepo, newID)

	getAccountsQuery := getaccounts.NewQuery(accountRepo)
	getActorsQuery := getactors.NewQuery(actorRepo, accountRepo)

	createAccountHTTP := createAccountEndpoint.NewHandler(createAccountHandler)
	createActorHTTP := createActorEndpoint.NewHandler(createActorHandler)
	getAccountsHTTP := getAccountsEndpoint.NewHandler(getAccountsQuery)
	getActorsHTTP := getActorsEndpoint.NewHandler(getActorsQuery)

	return &application{
		CreateAccountHandler: createAccountHandler,
		CreateActorHandler:   createActorHandler,
		CreateActorHTTP:      createActorHTTP,
		CreateAccountHTTP:    createAccountHTTP,
		GetAccountsHTTP:      getAccountsHTTP,
		GetActorsHTTP:        getActorsHTTP,
	}, nil
}

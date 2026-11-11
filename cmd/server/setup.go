package main

import (
	"uuid"

	"exercise/internal/participants/adapters/db"
	"exercise/internal/participants/adapters/db/sqlcgen"
	createAccountEndpoint "exercise/internal/participants/adapters/http/createaccount"
	createActorEndpoint "exercise/internal/participants/adapters/http/createactor"
	getAccountsEndpoint "exercise/internal/participants/adapters/http/getaccounts"
	getActorsEndpoint "exercise/internal/participants/adapters/http/getactors"
	createAccountCommand "exercise/internal/participants/app/command/createaccount"
	createActorCommand "exercise/internal/participants/app/command/createactor"
	"exercise/internal/participants/app/query/getaccounts"
	"exercise/internal/participants/app/query/getactors"

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

func setupApplication(pool *pgxpool.Pool) *application {
	q := sqlcgen.New(pool)
	accountRepo := db.NewAccountRepository(q)
	actorRepo := db.NewActorRepository(q)

	newID := uuid.NewV7
	createAccountHandler := createAccountCommand.NewHandler(db.NewUnitOfWork(pool), newID)
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
	}
}

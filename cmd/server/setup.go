package main

import (
	"uuid"

	"exercise/internal/participants/adapters/db"
	"exercise/internal/participants/adapters/db/sqlcgen"
	createAccountEndpoint "exercise/internal/participants/adapters/http/createaccount"
	createActorEndpoint "exercise/internal/participants/adapters/http/createactor"
	deleteActorEndpoint "exercise/internal/participants/adapters/http/deleteactor"
	getAccountsEndpoint "exercise/internal/participants/adapters/http/getaccounts"
	getActorsEndpoint "exercise/internal/participants/adapters/http/getactors"
	updateActorEndpoint "exercise/internal/participants/adapters/http/updateactor"
	"exercise/internal/participants/app/command/createaccount"
	"exercise/internal/participants/app/command/createactor"
	"exercise/internal/participants/app/command/deleteactor"
	"exercise/internal/participants/app/command/updateactor"
	"exercise/internal/participants/app/query/getaccounts"
	"exercise/internal/participants/app/query/getactors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type application struct {
	CreateActorHandler   *createactor.Handler
	CreateAccountHandler *createaccount.Handler
	GetAccountsQuery     *getaccounts.Query
	GetActorsQuery       *getactors.Query
	CreateAccountHTTP    *createAccountEndpoint.Handler
	CreateActorHTTP      *createActorEndpoint.Handler
	UpdateActorHTTP      *updateActorEndpoint.Handler
	DeleteActorHTTP      *deleteActorEndpoint.Handler
	GetAccountsHTTP      *getAccountsEndpoint.Handler
	GetActorsHTTP        *getActorsEndpoint.Handler
}

func setupApplication(pool *pgxpool.Pool) *application {
	q := sqlcgen.New(pool)
	accountRepo := db.NewAccountRepository(q)
	actorRepo := db.NewActorRepository(q)

	newID := uuid.NewV7
	createAccountHandler := createaccount.NewHandler(db.NewUnitOfWork(pool), newID)
	createActorHandler := createactor.NewHandler(actorRepo, newID)
	updateActorHandler := updateactor.NewHandler(actorRepo)
	deleteActorHandler := deleteactor.NewHandler(actorRepo)

	getAccountsQuery := getaccounts.NewQuery(accountRepo)
	getActorsQuery := getactors.NewQuery(actorRepo, accountRepo)

	createAccountHTTP := createAccountEndpoint.NewHandler(createAccountHandler)
	createActorHTTP := createActorEndpoint.NewHandler(createActorHandler)
	updateActorHTTP := updateActorEndpoint.NewHandler(updateActorHandler)
	deleteActorHTTP := deleteActorEndpoint.NewHandler(deleteActorHandler)
	getAccountsHTTP := getAccountsEndpoint.NewHandler(getAccountsQuery)
	getActorsHTTP := getActorsEndpoint.NewHandler(getActorsQuery)

	return &application{
		CreateAccountHandler: createAccountHandler,
		CreateActorHandler:   createActorHandler,
		CreateActorHTTP:      createActorHTTP,
		CreateAccountHTTP:    createAccountHTTP,
		UpdateActorHTTP:      updateActorHTTP,
		DeleteActorHTTP:      deleteActorHTTP,
		GetAccountsHTTP:      getAccountsHTTP,
		GetActorsHTTP:        getActorsHTTP,
	}
}

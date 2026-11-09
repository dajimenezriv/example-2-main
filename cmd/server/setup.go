package main

import (
	"database/sql"
	createAccountEndpoint "exercise/internal/adapters/http/createaccount"
	createActorEndpoint "exercise/internal/adapters/http/createactor"
	getAccountsEndpoint "exercise/internal/adapters/http/getaccounts"
	getActorsEndpoint "exercise/internal/adapters/http/getactors"
	accountrepo "exercise/internal/adapters/postgres/account"
	actorrepo "exercise/internal/adapters/postgres/actor"
	createAccountCommand "exercise/internal/app/command/createaccount"
	createActorCommand "exercise/internal/app/command/createactor"
	"exercise/internal/app/query/getaccounts"
	"exercise/internal/app/query/getactors"

	"uuid"
)

type context struct {
	CreateActorHandler   *createActorCommand.Handler
	CreateAccountHandler *createAccountCommand.Handler
	GetAccountsQuery     *getaccounts.Query
	GetActorsQuery       *getactors.Query
	CreateAccountHTTP    *createAccountEndpoint.Handler
	CreateActorHTTP      *createActorEndpoint.Handler
	GetAccountsHTTP      *getAccountsEndpoint.Handler
	GetActorsHTTP        *getActorsEndpoint.Handler
}

func setupContext(db *sql.DB) (*context, error) {
	accountRepo := accountrepo.NewPostgresRepository(db)
	actorRepo := actorrepo.NewPostgresRepository(db)

	newID := func() uuid.UUID { return uuid.NewV7() }
	createAccountHandler := createAccountCommand.NewHandler(accountRepo, actorRepo, newID)
	createActorHandler := createActorCommand.NewHandler(actorRepo, newID)

	getAccountsQuery := getaccounts.NewQuery(accountRepo)
	getActorsQuery := getactors.NewQuery(actorRepo, accountRepo)

	createAccountHTTP := createAccountEndpoint.NewHandler(createAccountHandler)
	createActorHTTP := createActorEndpoint.NewHandler(createActorHandler)
	getAccountsHTTP := getAccountsEndpoint.NewHandler(getAccountsQuery)
	getActorsHTTP := getActorsEndpoint.NewHandler(getActorsQuery)

	return &context{
		CreateAccountHandler: createAccountHandler,
		CreateActorHandler:   createActorHandler,
		CreateActorHTTP:      createActorHTTP,
		CreateAccountHTTP:    createAccountHTTP,
		GetAccountsHTTP:      getAccountsHTTP,
		GetActorsHTTP:        getActorsHTTP,
	}, nil
}

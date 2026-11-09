package main

import (
	"database/sql"
	createAccountCommand "exercise/internal/app/command/createaccount"
	createActorCommand "exercise/internal/app/command/createactor"
	"exercise/internal/app/query/getaccounts"
	"exercise/internal/app/query/getactors"
	createAccountEndpoint "exercise/internal/infra/http/createaccount"
	createActorEndpoint "exercise/internal/infra/http/createactor"
	getAccountsEndpoint "exercise/internal/infra/http/getaccounts"
	getActorsEndpoint "exercise/internal/infra/http/getactors"
	accountrepo "exercise/internal/infra/repository/account"
	actorrepo "exercise/internal/infra/repository/actor"

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

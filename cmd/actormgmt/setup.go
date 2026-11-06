package actormgmt

import (
	"database/sql"
	createAccountCommand "exercise2/cmd/actormgmt/internal/app/command/createaccount"
	createActorCommand "exercise2/cmd/actormgmt/internal/app/command/createactor"
	"exercise2/cmd/actormgmt/internal/app/query/getaccounts"
	"exercise2/cmd/actormgmt/internal/app/query/getactors"
	createAccountEndpoint "exercise2/cmd/actormgmt/internal/infra/http/createaccount"
	createActorEndpoint "exercise2/cmd/actormgmt/internal/infra/http/createactor"
	getAccountsEndpoint "exercise2/cmd/actormgmt/internal/infra/http/getaccounts"
	getActorsEndpoint "exercise2/cmd/actormgmt/internal/infra/http/getactors"
	accountrepo "exercise2/cmd/actormgmt/internal/infra/repository/account"
	"exercise2/cmd/actormgmt/internal/infra/repository/actor"
)

type Context struct {
	CreateActorHandler   *createActorCommand.Handler
	CreateAccountHandler *createAccountCommand.Handler
	GetAccountsQuery     *getaccounts.Query
	GetActorsQuery       *getactors.Query
	CreateAccountHTTP    *createAccountEndpoint.Handler
	CreateActorHTTP      *createActorEndpoint.Handler
	GetAccountsHTTP      *getAccountsEndpoint.Handler
	GetActorsHTTP        *getActorsEndpoint.Handler
}

func SetupContext(db *sql.DB) (*Context, error) {
	accountRepo := accountrepo.NewPostgresRepository(db)
	actorRepo := actorrepo.NewPostgresRepository(db)

	createAccountHandler := createAccountCommand.NewHandler(accountRepo)
	createActorHandler := createActorCommand.NewHandler(actorRepo)

	getAccountsQuery := getaccounts.NewQuery(accountRepo)
	getActorsQuery := getactors.NewQuery(actorRepo)

	createAccountHTTP := createAccountEndpoint.NewHandler(createAccountHandler)
	createActorHTTP := createActorEndpoint.NewHandler(createActorHandler)
	getAccountsHTTP := getAccountsEndpoint.NewHandler(getAccountsQuery)
	getActorsHTTP := getActorsEndpoint.NewHandler(getActorsQuery)

	return &Context{
		CreateAccountHandler: createAccountHandler,
		CreateActorHandler:   createActorHandler,
		CreateActorHTTP:      createActorHTTP,
		CreateAccountHTTP:    createAccountHTTP,
		GetAccountsHTTP:      getAccountsHTTP,
		GetActorsHTTP:        getActorsHTTP,
	}, nil
}

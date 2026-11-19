package main

import (
	"fmt"

	"exercise/internal/cli"
	"exercise/internal/cli/adapters/api"
	"exercise/internal/cli/adapters/auth0"
	createActorCommand "exercise/internal/cli/adapters/commands/createactor"
	listActorsCommand "exercise/internal/cli/adapters/commands/listactors"
	loginCommand "exercise/internal/cli/adapters/commands/login"
	"exercise/internal/cli/adapters/filestore"
)

type application struct {
	LoginCmd       *loginCommand.Command
	ListActorsCmd  *listActorsCommand.Command
	CreateActorCmd *createActorCommand.Command
}

func setupApplication(cfg *cli.Config) (*application, error) {
	auth0Cfg := auth0.Config{Issuer: cfg.Issuer, Audience: cfg.Audience, ClientID: cfg.ClientID}
	tokens, err := filestore.DefaultTokenStore()
	if err != nil {
		return nil, fmt.Errorf("token store: %w", err)
	}
	client, err := api.NewClient(cfg.APIURL, auth0.NewHTTPClient(auth0Cfg, tokens))
	if err != nil {
		return nil, fmt.Errorf("api client: %w", err)
	}
	actorRepo := api.NewActorRepository(client)

	return &application{
		LoginCmd:       loginCommand.NewCommand(auth0.NewAuthenticator(auth0Cfg), tokens),
		ListActorsCmd:  listActorsCommand.NewCommand(actorRepo),
		CreateActorCmd: createActorCommand.NewCommand(actorRepo),
	}, nil
}

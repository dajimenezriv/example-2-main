package api

import (
	"context"
	"fmt"
	"uuid"

	"exercise/internal/cli/adapters/api/apiclient"
	"exercise/internal/cli/domain/actor"
)

type ActorRepository struct {
	api *apiclient.ClientWithResponses
}

func NewActorRepository(api *apiclient.ClientWithResponses) *ActorRepository {
	return &ActorRepository{api: api}
}

func (r *ActorRepository) FindAll(ctx context.Context, limit, offset int32) (actor.Page, error) {
	resp, err := checkResponse(r.api.ListActorsWithResponse(ctx, &apiclient.ListActorsParams{Limit: &limit, Offset: &offset}))
	if err != nil {
		return actor.Page{}, fmt.Errorf("GET /actors: %w", err)
	}
	body, err := requireBody(resp.JSON200)
	if err != nil {
		return actor.Page{}, fmt.Errorf("GET /actors: %w", err)
	}

	items := derefSlice(body.Actors)
	page := actor.Page{
		Actors:      make([]actor.Actor, len(items)),
		TotalCount:  body.TotalCount,
		HasNext:     body.HasNext,
		HasPrevious: body.HasPrevious,
	}
	for i, item := range items {
		page.Actors[i] = toActorDomain(item)
	}
	return page, nil
}

func (r *ActorRepository) Create(ctx context.Context, taxID string, actorType actor.Type) (actor.Actor, error) {
	resp, err := checkResponse(r.api.CreateActorWithResponse(ctx, apiclient.CreateActorRequest{
		TaxId: taxID,
		Type:  apiclient.CreateActorRequestType(actorType),
	}))
	if err != nil {
		return actor.Actor{}, fmt.Errorf("POST /actors: %w", err)
	}
	body, err := requireBody(resp.JSON201)
	if err != nil {
		return actor.Actor{}, fmt.Errorf("POST /actors: %w", err)
	}
	return actor.Actor{
		ID:      uuid.UUID(body.Id),
		TaxID:   body.TaxId,
		Type:    actor.Type(body.Type),
		Version: body.Version,
	}, nil
}

func toActorDomain(item apiclient.ActorItem) actor.Actor {
	accountItems := derefSlice(item.Accounts)
	accounts := make([]actor.Account, len(accountItems))
	for i, acc := range accountItems {
		accounts[i] = actor.Account{ID: uuid.UUID(acc.Id), DltAccountID: acc.DltAccountId, PublicKey: acc.PublicKey}
	}
	return actor.Actor{
		// The generated client uses github.com/google/uuid; same [16]byte layout as stdlib uuid.
		ID:       uuid.UUID(item.Id),
		TaxID:    item.TaxId,
		Type:     actor.Type(item.Type),
		Version:  item.Version,
		Accounts: accounts,
	}
}

// The generated models use *[]T for arrays the spec marks nullable.
func derefSlice[T any](s *[]T) []T {
	if s == nil {
		return nil
	}
	return *s
}

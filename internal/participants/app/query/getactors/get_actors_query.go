package getactors

import (
	"context"
	"fmt"
	"uuid"

	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/pagination"
)

type ActorDTO struct {
	ID       uuid.UUID
	TaxID    actor.TaxID
	Type     actor.Type
	Version  int64
	Accounts []AccountDTO
}

type AccountDTO struct {
	ID           uuid.UUID
	DltAccountID string
	PublicKey    string
}

type Result struct {
	Actors      []ActorDTO
	TotalCount  int64
	HasNext     bool
	HasPrevious bool
}

type Query struct {
	RepoActor   actor.Repository
	RepoAccount account.Repository
}

func NewQuery(actorRepo actor.Repository, accountRepo account.Repository) *Query {
	return &Query{RepoActor: actorRepo, RepoAccount: accountRepo}
}

func (q *Query) Execute(ctx context.Context, limit, offset int32) (*Result, error) {
	page, err := pagination.NewPage(limit, offset, actor.MaxPageLimit)
	if err != nil {
		return nil, fmt.Errorf("build page: %w", err)
	}
	actors, err := q.RepoActor.FindAll(ctx, page)
	if err != nil {
		return nil, fmt.Errorf("find actors: %w", err)
	}
	total, err := q.RepoActor.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count actors: %w", err)
	}

	actorsIDs := make([]uuid.UUID, len(actors))
	for i, act := range actors {
		actorsIDs[i] = act.ID()
	}
	accounts, err := q.RepoAccount.FindByActorIDs(ctx, actorsIDs)
	if err != nil {
		return nil, fmt.Errorf("find accounts by actor IDs: %w", err)
	}

	byActor := make(map[uuid.UUID][]AccountDTO, len(actors))
	for _, acc := range accounts {
		byActor[acc.ActorID()] = append(byActor[acc.ActorID()], AccountDTO{
			ID:           acc.ID(),
			DltAccountID: acc.DltAccountID(),
			PublicKey:    acc.PublicKey(),
		})
	}

	actorDTOs := make([]ActorDTO, len(actors))
	for i, act := range actors {
		actorDTOs[i] = ActorDTO{
			ID:       act.ID(),
			TaxID:    act.TaxID(),
			Type:     act.Type(),
			Version:  act.Version(),
			Accounts: byActor[act.ID()],
		}
	}
	return &Result{
		Actors:      actorDTOs,
		TotalCount:  total,
		HasNext:     page.HasNext(total),
		HasPrevious: page.HasPrevious(),
	}, nil
}

package getactors

import (
	"context"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/actor"
	"uuid"
)

type ActorDTO struct {
	ID       uuid.UUID
	TaxID    string
	Type     actor.Type
	Accounts []AccountDTO
}

type AccountDTO struct {
	ID           uuid.UUID
	DltAccountID string
	PublicKey    string
}

type Query struct {
	RepoActor   actor.Repository
	RepoAccount account.Repository
}

func NewQuery(actorRepo actor.Repository, accountRepo account.Repository) *Query {
	return &Query{RepoActor: actorRepo, RepoAccount: accountRepo}
}

func (q *Query) Execute(ctx context.Context) ([]ActorDTO, error) {
	actors, err := q.RepoActor.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	actorsIDs := make([]uuid.UUID, len(actors))
	for i, act := range actors {
		actorsIDs[i] = act.ID
	}
	accounts, err := q.RepoAccount.FindByActorIDs(ctx, actorsIDs)
	if err != nil {
		return nil, err
	}

	byActor := make(map[uuid.UUID][]AccountDTO, len(actors))
	for _, acc := range accounts {
		byActor[acc.ActorID] = append(byActor[acc.ActorID], AccountDTO{
			ID:           acc.ID,
			DltAccountID: acc.DltAccountID,
			PublicKey:    acc.PublicKey,
		})
	}

	actorDTOs := make([]ActorDTO, len(actors))
	for i, act := range actors {
		actorDTOs[i] = ActorDTO{
			ID:       act.ID,
			TaxID:    act.TaxID,
			Type:     act.Type,
			Accounts: byActor[act.ID],
		}
	}
	return actorDTOs, nil
}

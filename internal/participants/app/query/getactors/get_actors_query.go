package getactors

import (
	"context"
	domainAccount "exercise/internal/participants/domain/account"
	domainActor "exercise/internal/participants/domain/actor"
	"uuid"
)

type ActorDTO struct {
	ID       uuid.UUID
	TaxId    string
	Type     domainActor.ActorType
	Accounts []AccountDTO
}

type AccountDTO struct {
	ID           uuid.UUID
	DltAccountId string
	PublicKey    string
}

type Query struct {
	RepoActor   domainActor.Repository
	RepoAccount domainAccount.Repository
}

func NewQuery(actorRepo domainActor.Repository, accountRepo domainAccount.Repository) *Query {
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
			DltAccountId: acc.DltAccountId,
			PublicKey:    acc.PublicKey,
		})
	}

	actorDTOs := make([]ActorDTO, len(actors))
	for i, act := range actors {
		actorDTOs[i] = ActorDTO{
			ID:       act.ID,
			TaxId:    act.TaxId,
			Type:     act.Type,
			Accounts: byActor[act.ID],
		}
	}
	return actorDTOs, nil
}

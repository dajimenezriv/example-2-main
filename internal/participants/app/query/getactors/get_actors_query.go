package getactors

import (
	"context"
	domainAccount "exercise/internal/participants/domain/account"
	domainActor "exercise/internal/participants/domain/actor"
	"uuid"
)

type Query struct {
	RepoActor   domainActor.Repository
	RepoAccount domainAccount.Repository
}

func NewQuery(actor domainActor.Repository, account domainAccount.Repository) *Query {
	return &Query{RepoActor: actor, RepoAccount: account}
}

func (q *Query) Execute(ctx context.Context) ([]*ActorDTO, error) {
	actors, err := q.RepoActor.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	actorsIDs := make([]uuid.UUID, len(actors))
	for i, a := range actors {
		actorsIDs[i] = a.ID
	}
	accounts, err := q.RepoAccount.FindByActorIDs(ctx, actorsIDs)
	if err != nil {
		return nil, err
	}

	var actDTOs []*ActorDTO
	for _, a := range actors {
		accDTOs := make([]AccountDTO, len(accounts))
		for i, acc := range accounts {
			if acc.ActorID != a.ID {
				continue
			}
			accDTOs[i] = AccountDTO{
				ID:           acc.ID.String(),
				DltAccountId: acc.DltAccountId,
				PublicKey:    acc.PublicKey,
			}
		}

		actDTOs = append(actDTOs, &ActorDTO{
			ID:       a.ID.String(),
			TaxId:    a.TaxId,
			Type:     string(a.Type),
			Accounts: accDTOs,
		})
	}
	return actDTOs, nil
}

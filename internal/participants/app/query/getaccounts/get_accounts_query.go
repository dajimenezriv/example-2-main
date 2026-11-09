package getaccounts

import (
	"context"
	domainAccount "exercise/internal/participants/domain/account"
)

type Query struct {
	Repo domainAccount.Repository
}

func NewQuery(repo domainAccount.Repository) *Query {
	return &Query{Repo: repo}
}

func (q *Query) Execute(ctx context.Context) ([]*AccountDTO, error) {
	accounts, err := q.Repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var dtos []*AccountDTO
	for _, a := range accounts {
		dtos = append(dtos, &AccountDTO{
			ID:           a.ID.String(),
			DltAccountId: a.DltAccountId,
			PrivateKey:   a.PrivateKey,
			PublicKey:    a.PublicKey,
		})
	}
	return dtos, nil
}

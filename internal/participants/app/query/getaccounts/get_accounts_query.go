package getaccounts

import (
	"context"
	"fmt"

	domainAccount "exercise/internal/participants/domain/account"
)

type AccountDTO struct {
	ID           string
	DltAccountID string
	PrivateKey   string
	PublicKey    string
	Version      int64
}

type Query struct {
	Repo domainAccount.Repository
}

func NewQuery(repo domainAccount.Repository) *Query {
	return &Query{Repo: repo}
}

func (q *Query) Execute(ctx context.Context, limit, offset int) ([]*AccountDTO, error) {
	accounts, err := q.Repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find accounts: %w", err)
	}

	var dtos []*AccountDTO
	for _, a := range accounts {
		dtos = append(dtos, &AccountDTO{
			ID:           a.ID().String(),
			DltAccountID: a.DltAccountID(),
			PrivateKey:   a.PrivateKey(),
			PublicKey:    a.PublicKey(),
			Version:      a.Version(),
		})
	}
	return dtos, nil
}

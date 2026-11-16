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

type Result struct {
	Accounts    []*AccountDTO
	TotalCount  int64
	HasNext     bool
	HasPrevious bool
}

func NewQuery(repo domainAccount.Repository) *Query {
	return &Query{Repo: repo}
}

func (q *Query) Execute(ctx context.Context, limit, offset int) (*Result, error) {
	accounts, err := q.Repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find accounts: %w", err)
	}
	total, err := q.Repo.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count accounts: %w", err)
	}

	res := &Result{
		Accounts:    make([]*AccountDTO, len(accounts)),
		TotalCount:  total,
		HasNext:     int64(offset+len(accounts)) < total,
		HasPrevious: offset > 0,
	}
	for i, a := range accounts {
		res.Accounts[i] = &AccountDTO{
			ID:           a.ID().String(),
			DltAccountID: a.DltAccountID(),
			PrivateKey:   a.PrivateKey(),
			PublicKey:    a.PublicKey(),
			Version:      a.Version(),
		}
	}
	return res, nil
}

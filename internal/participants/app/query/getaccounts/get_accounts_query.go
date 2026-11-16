package getaccounts

import (
	"context"
	"fmt"

	"exercise/internal/participants/domain/account"
)

type AccountDTO struct {
	ID           string
	DltAccountID string
	PrivateKey   string
	PublicKey    string
	Version      int64
}

type Query struct {
	Repo account.Repository
}

type Result struct {
	Accounts    []*AccountDTO
	TotalCount  int64
	HasNext     bool
	HasPrevious bool
}

func NewQuery(repo account.Repository) *Query {
	return &Query{Repo: repo}
}

func (q *Query) Execute(ctx context.Context, limit, offset int32) (*Result, error) {
	page, err := account.NewPage(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("build page: %w", err)
	}
	accounts, err := q.Repo.FindAll(ctx, page)
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
		HasNext:     page.HasNext(total),
		HasPrevious: page.HasPrevious(),
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

package uow

import (
	"context"

	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/actor"
)

type Repositories struct {
	Accounts account.Repository
	Actors   actor.Repository
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context, repos Repositories) error) error
}

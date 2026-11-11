package uowtest

import (
	"context"

	"exercise/internal/participants/app/uow"
)

type fakeUnitOfWork struct {
	repos uow.Repositories
}

func NewFake(repos uow.Repositories) uow.UnitOfWork {
	return &fakeUnitOfWork{repos: repos}
}

func (f *fakeUnitOfWork) Do(ctx context.Context, fn func(ctx context.Context, repos uow.Repositories) error) error {
	return fn(ctx, f.repos)
}

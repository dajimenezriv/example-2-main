package account

import (
	"context"
	"uuid"

	"exercise/internal/participants/domain/pagination"
)

type Repository interface {
	Count(ctx context.Context) (int64, error)
	Save(ctx context.Context, account *Account) error
	FindByID(ctx context.Context, id uuid.UUID) (*Account, error)
	FindAll(ctx context.Context, page pagination.Page) ([]*Account, error)
	FindByActorIDs(ctx context.Context, actorIDs []uuid.UUID) ([]*Account, error)
}

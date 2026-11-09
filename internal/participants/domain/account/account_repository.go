package account

import (
	"context"
	"uuid"
)

type Repository interface {
	Save(ctx context.Context, account *Account) error
	FindByID(ctx context.Context, id uuid.UUID) (*Account, error)
	FindAll(ctx context.Context) ([]*Account, error)
	FindByActorIDs(ctx context.Context, actorIDs []uuid.UUID) ([]*Account, error)
}

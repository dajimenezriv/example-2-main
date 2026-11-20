package actor

import (
	"context"
	"uuid"

	"exercise/internal/participants/domain/pagination"
)

type Repository interface {
	Count(ctx context.Context) (int64, error)
	Save(ctx context.Context, actor *Actor) error
	Update(ctx context.Context, actor *Actor, version int64) (*Actor, error)
	Delete(ctx context.Context, id uuid.UUID, version int64) error
	FindByID(ctx context.Context, id uuid.UUID) (*Actor, error)
	FindAll(ctx context.Context, page pagination.Page) ([]*Actor, error)
}

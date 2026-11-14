package actor

import (
	"context"
	"uuid"
)

type Repository interface {
	Save(ctx context.Context, actor *Actor) error
	Update(ctx context.Context, actor *Actor, version int64) (*Actor, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Actor, error)
	FindAll(ctx context.Context) ([]*Actor, error)
}

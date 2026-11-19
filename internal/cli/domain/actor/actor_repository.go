package actor

import "context"

type Repository interface {
	FindAll(ctx context.Context, limit, offset int32) (Page, error)
	Create(ctx context.Context, taxID string, actorType Type) (Actor, error)
}

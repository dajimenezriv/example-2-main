package actortest

import (
	"context"
	"sync"
	"uuid"

	"exercise/internal/cli/domain/actor"
)

type fakeRepository struct {
	mu     sync.Mutex
	actors []actor.Actor
}

func NewFakeRepository(actors ...actor.Actor) actor.Repository {
	return &fakeRepository{actors: append([]actor.Actor{}, actors...)}
}

// Paginates like the API: HasPrevious whenever offset > 0, even past the end.
func (f *fakeRepository) FindAll(ctx context.Context, limit, offset int32) (actor.Page, error) {
	if err := ctx.Err(); err != nil {
		return actor.Page{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	total := int32(len(f.actors)) //nolint:gosec // G115: a test fake never holds 2^31 actors.
	start := min(offset, total)
	end := min(start+limit, total)
	return actor.Page{
		Actors:      append([]actor.Actor{}, f.actors[start:end]...),
		TotalCount:  int64(total),
		HasNext:     end < total,
		HasPrevious: offset > 0,
	}, nil
}

func (f *fakeRepository) Create(ctx context.Context, taxID string, actorType actor.Type) (actor.Actor, error) {
	if err := ctx.Err(); err != nil {
		return actor.Actor{}, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	a := actor.Actor{ID: uuid.NewV7(), TaxID: taxID, Type: actorType, Version: 1}
	f.actors = append(f.actors, a)
	return a, nil
}

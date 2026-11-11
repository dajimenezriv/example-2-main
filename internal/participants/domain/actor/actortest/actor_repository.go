package actortest

import (
	"context"
	"sync"
	"uuid"

	"exercise/internal/participants/domain/actor"
)

type fakeRepository struct {
	mu     sync.Mutex
	actors []actor.Actor
}

func NewFakeRepository() actor.Repository {
	return &fakeRepository{
		actors: []actor.Actor{},
	}
}

func (f *fakeRepository) Save(ctx context.Context, act *actor.Actor) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, a := range f.actors {
		if a.ID == act.ID {
			return actor.ErrActorAlreadyExists
		}
	}
	f.actors = append(f.actors, *act)
	return nil
}

func (f *fakeRepository) FindByID(ctx context.Context, id uuid.UUID) (*actor.Actor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, a := range f.actors {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, actor.ErrActorNotFound
}

func (f *fakeRepository) FindAll(ctx context.Context) ([]*actor.Actor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	accs := make([]*actor.Actor, len(f.actors))
	for i, acc := range f.actors {
		accs[i] = &acc
	}
	return accs, nil
}

package actortest

import (
	"exercise2/internal/domain/actor"
	"sync"

	"uuid"
)

type FakeRepository struct {
	mu     sync.Mutex
	Actors []actor.Actor
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		Actors: []actor.Actor{},
	}
}

func (f *FakeRepository) Save(act *actor.Actor) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, a := range f.Actors {
		if a.ID == act.ID {
			return actor.ErrActorAlreadyExists
		}
	}
	f.Actors = append(f.Actors, *act)
	return nil
}

func (f *FakeRepository) FindByID(id uuid.UUID) (*actor.Actor, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, a := range f.Actors {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, actor.ErrActorNotFound
}

func (f *FakeRepository) FindAll() ([]*actor.Actor, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	accs := make([]*actor.Actor, len(f.Actors))
	for i, acc := range f.Actors {
		accs[i] = &acc
	}
	return accs, nil
}

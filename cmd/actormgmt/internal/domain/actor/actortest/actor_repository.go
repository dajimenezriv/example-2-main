package actortest

import (
	"exercise2/cmd/actormgmt/internal/domain/actor"
	actorrepo "exercise2/cmd/actormgmt/internal/infra/repository/actor"
	"sync"

	"uuid"
)

type FakeRepository struct {
	mu     sync.Mutex
	Actors map[uuid.UUID]*actorrepo.Actor

	SaveErr, FindByIDErr, FindAllErr error
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		Actors: make(map[uuid.UUID]*actorrepo.Actor),
	}
}

func (f *FakeRepository) Save(actor *actor.Actor) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Actors[actor.ID] = actorrepo.ToPersistenceModel(actor)
	return f.SaveErr
}

func (f *FakeRepository) FindByID(id uuid.UUID) (*actor.Actor, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if acc, ok := f.Actors[id]; ok {
		return actorrepo.ToDomain(acc), f.FindByIDErr
	}
	return nil, actor.ErrActorNotFound
}

func (f *FakeRepository) FindAll() ([]*actor.Actor, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var accs []*actor.Actor
	for _, acc := range f.Actors {
		accs = append(accs, actorrepo.ToDomain(acc))
	}
	return accs, f.FindAllErr
}

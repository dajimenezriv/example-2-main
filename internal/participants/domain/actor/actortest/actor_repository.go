package actortest

import (
	"context"
	"slices"
	"sync"
	"uuid"

	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/pagination"
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

func (f *fakeRepository) Count(ctx context.Context) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	return int64(len(f.actors)), nil
}

func (f *fakeRepository) Save(ctx context.Context, act *actor.Actor) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, a := range f.actors {
		if a.ID() == act.ID() {
			return actor.ErrAlreadyExists
		}
		if a.TaxID() == act.TaxID() {
			return actor.ErrTaxIDAlreadyExists
		}
	}
	f.actors = append(f.actors, *act)
	return nil
}

func (f *fakeRepository) Update(ctx context.Context, act *actor.Actor, expectedVersion int64) (*actor.Actor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	idx := -1
	for i, a := range f.actors {
		if a.ID() == act.ID() {
			if a.Version() != expectedVersion {
				return nil, actor.ErrVersionConflict
			}
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, actor.ErrNotFound
	}
	for i, a := range f.actors {
		if i == idx {
			continue
		}
		if a.TaxID() == act.TaxID() {
			return nil, actor.ErrTaxIDAlreadyExists
		}
	}
	act = actor.Reconstitute(act.ID(), string(act.TaxID()), act.Type(), expectedVersion+1)
	f.actors[idx] = *act
	return act, nil
}

func (f *fakeRepository) Delete(ctx context.Context, id uuid.UUID, expectedVersion int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	for i, a := range f.actors {
		if a.ID() == id {
			if a.Version() != expectedVersion {
				return actor.ErrVersionConflict
			}
			f.actors = slices.Delete(f.actors, i, i+1)
			return nil
		}
	}
	return actor.ErrNotFound
}

func (f *fakeRepository) FindByID(ctx context.Context, id uuid.UUID) (*actor.Actor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, a := range f.actors {
		if a.ID() == id {
			return &a, nil
		}
	}
	return nil, actor.ErrNotFound
}

func (f *fakeRepository) FindAll(ctx context.Context, page pagination.Page) ([]*actor.Actor, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	from := min(len(f.actors), int(page.Offset()))
	to := min(len(f.actors), from+int(page.Limit()))
	acts := make([]*actor.Actor, to-from)

	for i, act := range f.actors[from:to] {
		acts[i] = &act
	}
	return acts, nil
}

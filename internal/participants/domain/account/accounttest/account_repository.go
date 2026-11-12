package accounttest

import (
	"context"
	"slices"
	"sync"
	"uuid"

	"exercise/internal/participants/domain/account"
)

type fakeRepository struct {
	mu       sync.Mutex
	accounts []account.Account
}

func NewFakeRepository() account.Repository {
	return &fakeRepository{
		accounts: []account.Account{},
	}
}

func (f *fakeRepository) Save(ctx context.Context, acc *account.Account) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, a := range f.accounts {
		if a.ID() == acc.ID() {
			return account.ErrAlreadyExists
		}
	}
	f.accounts = append(f.accounts, *acc)
	return nil
}

func (f *fakeRepository) FindByID(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, a := range f.accounts {
		if a.ID() == id {
			return &a, nil
		}
	}
	return nil, account.ErrNotFound
}

func (f *fakeRepository) FindAll(ctx context.Context) ([]*account.Account, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	accs := make([]*account.Account, len(f.accounts))
	for i, acc := range f.accounts {
		accs[i] = &acc
	}
	return accs, nil
}

func (f *fakeRepository) FindByActorIDs(ctx context.Context, actorIDs []uuid.UUID) ([]*account.Account, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	var accs []*account.Account
	for _, acc := range f.accounts {
		if slices.Contains(actorIDs, acc.ActorID()) {
			accs = append(accs, &acc)
		}
	}
	return accs, nil
}

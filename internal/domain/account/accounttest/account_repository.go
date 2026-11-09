package accounttest

import (
	"exercise/internal/domain/account"
	"slices"
	"sync"

	"uuid"
)

type FakeRepository struct {
	mu       sync.Mutex
	Accounts []account.Account
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		Accounts: []account.Account{},
	}
}

func (f *FakeRepository) Save(acc *account.Account) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, a := range f.Accounts {
		if a.ID == acc.ID {
			return account.ErrAccountAlreadyExists
		}
	}
	f.Accounts = append(f.Accounts, *acc)
	return nil
}

func (f *FakeRepository) FindByID(id uuid.UUID) (*account.Account, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, a := range f.Accounts {
		if a.ID == id {
			return &a, nil
		}
	}
	return nil, account.ErrAccountNotFound
}

func (f *FakeRepository) FindAll() ([]*account.Account, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	accs := make([]*account.Account, len(f.Accounts))
	for i, acc := range f.Accounts {
		accs[i] = &acc
	}
	return accs, nil
}

func (f *FakeRepository) FindByActorIDs(actorIDs []uuid.UUID) ([]*account.Account, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var accs []*account.Account
	for _, acc := range f.Accounts {
		if slices.Contains(actorIDs, acc.ActorID) {
			accs = append(accs, &acc)
		}
	}
	return accs, nil
}

package accounttest

import (
	"exercise2/cmd/actormgmt/internal/domain/account"
	accountrepo "exercise2/cmd/actormgmt/internal/infra/repository/account"
	"slices"
	"sync"

	"uuid"
)

type FakeRepository struct {
	mu       sync.Mutex
	Accounts map[uuid.UUID]*accountrepo.Account

	SaveErr, FindByIDErr, FindAllErr, FindByActorIDsErr error
}

func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		Accounts: make(map[uuid.UUID]*accountrepo.Account),
	}
}

func (f *FakeRepository) Save(account *account.Account) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Accounts[account.ID] = accountrepo.ToPersistenceModel(account)
	return f.SaveErr
}

func (f *FakeRepository) FindByID(id uuid.UUID) (*account.Account, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if acc, ok := f.Accounts[id]; ok {
		return accountrepo.ToDomain(acc), f.FindByIDErr
	}
	return nil, account.ErrAccountNotFound
}

func (f *FakeRepository) FindAll() ([]*account.Account, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var accs []*account.Account
	for _, acc := range f.Accounts {
		accs = append(accs, accountrepo.ToDomain(acc))
	}
	return accs, f.FindAllErr
}

func (f *FakeRepository) FindByActorIDs(actorIDs []uuid.UUID) ([]*account.Account, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var accs []*account.Account
	for _, acc := range f.Accounts {
		if slices.Contains(actorIDs, acc.ActorID) {
			accs = append(accs, accountrepo.ToDomain(acc))
		}
	}
	return accs, f.FindByActorIDsErr
}

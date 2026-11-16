package accounttest

import (
	"context"
	"math"
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

func (f *fakeRepository) Count(ctx context.Context) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	return int64(len(f.accounts)), nil
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

func (f *fakeRepository) FindAll(ctx context.Context, limit, offset int) ([]*account.Account, error) {
	if limit < 0 || offset < 0 || limit > math.MaxInt32 || offset > math.MaxInt32 {
		return nil, account.ErrInvalidPagination
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	f.mu.Lock()
	defer f.mu.Unlock()

	from := min(len(f.accounts), offset)          // 0 ≤ from ≤ len
	to := from + min(len(f.accounts)-from, limit) // from ≤ to ≤ len

	accs := make([]*account.Account, to-from)
	for i, acc := range f.accounts[from:to] {
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

package db

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"exercise/internal/participants/adapters/db/sqlcgen"
	"exercise/internal/participants/domain/account"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type AccountRepository struct {
	db *sqlcgen.Queries
}

func NewAccountRepository(db *sqlcgen.Queries) *AccountRepository {
	return &AccountRepository{db: db}
}

func toAccountDomain(d sqlcgen.Account) *account.Account {
	return account.Reconstitute(
		d.ID,
		d.ActorID,
		d.DltAccountID,
		d.PrivateKey,
		d.PublicKey,
		d.Version,
	)
}

func (r *AccountRepository) Count(ctx context.Context) (int64, error) {
	n, err := r.db.CountAccounts(ctx)
	if err != nil {
		return 0, fmt.Errorf("count accounts: %w", err)
	}
	return n, nil
}

func (r *AccountRepository) Save(ctx context.Context, a *account.Account) error {
	err := r.db.CreateAccount(ctx, sqlcgen.CreateAccountParams{
		ID:           a.ID(),
		ActorID:      a.ActorID(),
		DltAccountID: a.DltAccountID(),
		PrivateKey:   a.PrivateKey(),
		PublicKey:    a.PublicKey(),
		Version:      a.Version(),
	})
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolation {
			return account.ErrAlreadyExists
		}
		return fmt.Errorf("save account: %w", err)
	}
	return nil
}

func (r *AccountRepository) FindByID(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	a, err := r.db.FindAccountByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, account.ErrNotFound
		}
		return nil, fmt.Errorf("find account by id: %w", err)
	}
	return toAccountDomain(a), nil
}

func (r *AccountRepository) FindAll(ctx context.Context, limit, offset int) ([]*account.Account, error) {
	if limit < 0 || offset < 0 {
		return nil, account.ErrInvalidPagination
	}

	as, err := r.db.FindAccounts(ctx, sqlcgen.FindAccountsParams{
		PageSize:   int32(limit),
		PageOffset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("find accounts: %w", err)
	}
	res := make([]*account.Account, len(as))
	for i, a := range as {
		res[i] = toAccountDomain(a)
	}
	return res, nil
}

func (r *AccountRepository) FindByActorIDs(ctx context.Context, actorIDs []uuid.UUID) ([]*account.Account, error) {
	if len(actorIDs) == 0 {
		return []*account.Account{}, nil
	}
	as, err := r.db.FindAccountsByActorIDs(ctx, actorIDs)
	if err != nil {
		return nil, fmt.Errorf("find accounts by actor ids: %w", err)
	}
	res := make([]*account.Account, len(as))
	for i, a := range as {
		res[i] = toAccountDomain(a)
	}
	return res, nil
}

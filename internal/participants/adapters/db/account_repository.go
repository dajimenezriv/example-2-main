package db

import (
	"context"
	"database/sql"
	"errors"

	"exercise/internal/participants/adapters/db/sqlcgen"
	"exercise/internal/participants/domain/account"
	"uuid"

	"github.com/jackc/pgx/v5/pgconn"
)

type AccountRepository struct {
	db *sqlcgen.Queries
}

func NewAccountRepository(db *sqlcgen.Queries) *AccountRepository {
	return &AccountRepository{db: db}
}

func toAccountDomain(d sqlcgen.Account) *account.Account {
	return &account.Account{
		ID:           d.ID,
		ActorID:      d.ActorID,
		DltAccountID: d.DltAccountID,
		PrivateKey:   d.PrivateKey,
		PublicKey:    d.PublicKey,
	}
}

func (r *AccountRepository) Save(ctx context.Context, a *account.Account) error {
	err := r.db.CreateAccount(ctx, sqlcgen.CreateAccountParams{
		ID:           a.ID,
		ActorID:      a.ActorID,
		DltAccountID: a.DltAccountID,
		PrivateKey:   a.PrivateKey,
		PublicKey:    a.PublicKey,
	})
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolation {
		return account.ErrAccountAlreadyExists
	}
	return err
}

func (r *AccountRepository) FindByID(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	a, err := r.db.FindAccountByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, account.ErrAccountNotFound
		}
		return nil, err
	}
	return toAccountDomain(a), nil
}

func (r *AccountRepository) FindAll(ctx context.Context) ([]*account.Account, error) {
	as, err := r.db.FindAccounts(ctx)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	res := make([]*account.Account, len(as))
	for i, a := range as {
		res[i] = toAccountDomain(a)
	}
	return res, nil
}

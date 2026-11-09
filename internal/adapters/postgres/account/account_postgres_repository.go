package accountrepo

import (
	"context"
	"database/sql"
	"errors"
	"uuid"

	"exercise/internal/adapters/postgres/sqlcgen"
	"exercise/internal/domain/account"
	"exercise/internal/domain/actor"

	"github.com/lib/pq"
)

const uniqueViolation = "23505"

type AccountPostgresRepository struct {
	db *sqlcgen.Queries
}

func NewAccountPostgresRepository(db *sqlcgen.Queries) *AccountPostgresRepository {
	return &AccountPostgresRepository{db: db}
}

func toAccountDomain(d sqlcgen.Account) *account.Account {
	return &account.Account{
		ID:           d.ID,
		ActorID:      d.ActorID,
		DltAccountId: d.DltAccountID,
		PrivateKey:   d.PrivateKey,
		PublicKey:    d.PublicKey,
	}
}

func (r *AccountPostgresRepository) Save(ctx context.Context, a *account.Account) error {
	err := r.db.CreateAccount(ctx, sqlcgen.CreateAccountParams{
		ID:           a.ID,
		ActorID:      a.ActorID,
		DltAccountID: a.DltAccountId,
		PrivateKey:   a.PrivateKey,
		PublicKey:    a.PublicKey,
	})
	if pgErr, ok := errors.AsType[*pq.Error](err); ok && pgErr.Code == uniqueViolation {
		return actor.ErrActorAlreadyExists
	}
	return err
}

func (r *AccountPostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	a, err := r.db.FindAccountByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, actor.ErrActorNotFound
	}
	return toAccountDomain(a), nil
}

func (r *AccountPostgresRepository) FindAll(ctx context.Context) ([]*account.Account, error) {
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

func (r *AccountPostgresRepository) FindByActorIDs(ctx context.Context, actorIDs []uuid.UUID) ([]*account.Account, error) {
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

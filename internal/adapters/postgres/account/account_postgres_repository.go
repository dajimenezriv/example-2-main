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

type PostgresRepository struct {
	db *sqlcgen.Queries
}

func NewPostgresRepository(db *sqlcgen.Queries) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(ctx context.Context, a *account.Account) error {
	model := ToPersistenceModel(a)
	_, err := r.db.CreateAccount(ctx, model)
	if pgErr, ok := errors.AsType[*pq.Error](err); ok && pgErr.Code == uniqueViolation {
		return actor.ErrActorAlreadyExists
	}
	return err
}

func (r *PostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*account.Account, error) {
	a, err := r.db.FindAccountByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, actor.ErrActorNotFound
	}
	return ToDomain(a), nil
}

func (r *PostgresRepository) FindAll(ctx context.Context) ([]*account.Account, error) {
	as, err := r.db.FindAccounts(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*account.Account, len(as))
	for i, a := range as {
		res[i] = ToDomain(a)
	}
	return res, nil
}

func (r *PostgresRepository) FindByActorIDs(ctx context.Context, actorIDs []uuid.UUID) ([]*account.Account, error) {
	as, err := r.db.FindAccountsByActorIDs(ctx, actorIDs)
	if err != nil {
		return nil, err
	}
	res := make([]*account.Account, len(as))
	for i, a := range as {
		res[i] = ToDomain(a)
	}
	return res, nil
}

package actorrepo

import (
	"context"
	"database/sql"
	"errors"

	"exercise/internal/domain/actor"
	"exercise/postgres"

	"uuid"

	"github.com/lib/pq"
)

const uniqueViolation = "23505"

type PostgresRepository struct {
	db *postgres.Queries
}

func NewPostgresRepository(db *postgres.Queries) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(ctx context.Context, a *actor.Actor) error {
	model := ToPersistenceModel(a)
	_, err := r.db.CreateActor(ctx, model)
	if pgErr, ok := errors.AsType[*pq.Error](err); ok && pgErr.Code == uniqueViolation {
		return actor.ErrActorAlreadyExists
	}
	return err
}

func (r *PostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*actor.Actor, error) {
	a, err := r.db.FindActorByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, actor.ErrActorNotFound
	}
	return ToDomain(a), nil
}

func (r *PostgresRepository) FindAll(ctx context.Context) ([]*actor.Actor, error) {
	as, err := r.db.FindActors(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*actor.Actor, len(as))
	for i, a := range as {
		res[i] = ToDomain(a)
	}
	return res, nil
}

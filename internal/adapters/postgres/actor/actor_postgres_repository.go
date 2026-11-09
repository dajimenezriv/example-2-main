package actorrepo

import (
	"context"
	"database/sql"
	"errors"

	"exercise/internal/adapters/postgres/sqlcgen"
	"exercise/internal/domain/actor"

	"uuid"

	"github.com/lib/pq"
)

const uniqueViolation = "23505"

type ActorPostgresRepository struct {
	db *sqlcgen.Queries
}

func NewActorPostgresRepository(db *sqlcgen.Queries) *ActorPostgresRepository {
	return &ActorPostgresRepository{db: db}
}

func toActorDomain(d sqlcgen.Actor) *actor.Actor {
	return &actor.Actor{
		ID:    d.ID,
		TaxId: d.TaxID,
		Type:  actor.ActorType(d.Type),
	}
}

func (r *ActorPostgresRepository) Save(ctx context.Context, a *actor.Actor) error {
	err := r.db.CreateActor(ctx, sqlcgen.CreateActorParams{
		ID:    a.ID,
		TaxID: a.TaxId,
		Type:  string(a.Type),
	})
	if pgErr, ok := errors.AsType[*pq.Error](err); ok && pgErr.Code == uniqueViolation {
		return actor.ErrActorAlreadyExists
	}
	return err
}

func (r *ActorPostgresRepository) FindByID(ctx context.Context, id uuid.UUID) (*actor.Actor, error) {
	a, err := r.db.FindActorByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, actor.ErrActorNotFound
	}
	return toActorDomain(a), nil
}

func (r *ActorPostgresRepository) FindAll(ctx context.Context) ([]*actor.Actor, error) {
	as, err := r.db.FindActors(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]*actor.Actor, len(as))
	for i, a := range as {
		res[i] = toActorDomain(a)
	}
	return res, nil
}

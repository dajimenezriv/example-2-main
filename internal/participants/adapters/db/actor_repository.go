package db

import (
	"context"
	"errors"
	"fmt"
	"uuid"

	"exercise/internal/participants/adapters/db/sqlcgen"
	"exercise/internal/participants/domain/actor"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ActorRepository struct {
	db *sqlcgen.Queries
}

func NewActorRepository(db *sqlcgen.Queries) *ActorRepository {
	return &ActorRepository{db: db}
}

func toActorDomain(d sqlcgen.Actor) *actor.Actor {
	return actor.Reconstitute(d.ID, d.TaxID, actor.Type(d.Type), d.Version)
}

func (r *ActorRepository) Save(ctx context.Context, a *actor.Actor) error {
	err := r.db.CreateActor(ctx, sqlcgen.CreateActorParams{
		ID:      a.ID(),
		TaxID:   string(a.TaxID()),
		Type:    string(a.Type()),
		Version: a.Version(),
	})
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolation {
			switch pgErr.ConstraintName {
			case actorsPKeyConstraint:
				return actor.ErrAlreadyExists
			case actorsTaxIDConstraint:
				return actor.ErrTaxIDAlreadyExists
			}
		}
		return fmt.Errorf("save actor: %w", err)
	}
	return nil
}

func (r *ActorRepository) Update(ctx context.Context, a *actor.Actor) (*actor.Actor, error) {
	d, err := r.db.UpdateActor(ctx, sqlcgen.UpdateActorParams{
		ID:      a.ID(),
		TaxID:   string(a.TaxID()),
		Type:    string(a.Type()),
		Version: a.Version(),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, actor.ErrNotFound
		}
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == uniqueViolation {
			switch pgErr.ConstraintName {
			case actorsTaxIDConstraint:
				return nil, actor.ErrTaxIDAlreadyExists
			}
		}
		return nil, fmt.Errorf("update actor: %w", err)
	}
	return toActorDomain(d), nil
}

func (r *ActorRepository) FindByID(ctx context.Context, id uuid.UUID) (*actor.Actor, error) {
	a, err := r.db.FindActorByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, actor.ErrNotFound
		}
		return nil, fmt.Errorf("find actor by id: %w", err)
	}
	return toActorDomain(a), nil
}

func (r *ActorRepository) FindAll(ctx context.Context) ([]*actor.Actor, error) {
	as, err := r.db.FindActors(ctx)
	if err != nil {
		return nil, fmt.Errorf("find actors: %w", err)
	}
	res := make([]*actor.Actor, len(as))
	for i, a := range as {
		res[i] = toActorDomain(a)
	}
	return res, nil
}

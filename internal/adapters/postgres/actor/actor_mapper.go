package actorrepo

import (
	"exercise/internal/adapters/postgres/sqlcgen"
	"exercise/internal/domain/actor"
)

func ToPersistenceModel(a *actor.Actor) sqlcgen.CreateActorParams {
	return sqlcgen.CreateActorParams{
		ID:    a.ID,
		TaxID: a.TaxId,
		Type:  string(a.Type),
	}
}

func ToDomain(d sqlcgen.Actor) *actor.Actor {
	return &actor.Actor{
		ID:    d.ID,
		TaxId: d.TaxID,
		Type:  actor.ActorType(d.Type),
	}
}

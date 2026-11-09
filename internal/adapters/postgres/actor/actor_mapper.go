package actorrepo

import (
	"exercise/internal/domain/actor"
	"exercise/postgres"
)

func ToPersistenceModel(a *actor.Actor) postgres.CreateActorParams {
	return postgres.CreateActorParams{
		ID:    a.ID,
		TaxID: a.TaxId,
		Type:  string(a.Type),
	}
}

func ToDomain(d postgres.Actor) *actor.Actor {
	return &actor.Actor{
		ID:    d.ID,
		TaxId: d.TaxID,
		Type:  actor.ActorType(d.Type),
	}
}

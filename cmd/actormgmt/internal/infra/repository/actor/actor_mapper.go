package actorrepo

import "exercise2/cmd/actormgmt/internal/domain/actor"

func ToPersistenceModel(a *actor.Actor) *Actor {
	return &Actor{
		ID:    a.ID,
		TaxId: a.TaxId,
		Type:  string(a.Type),
	}
}

func ToDomain(d *Actor) *actor.Actor {
	return &actor.Actor{
		ID:    d.ID,
		TaxId: d.TaxId,
		Type:  actor.ActorType(d.Type),
	}
}

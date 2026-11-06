package actor

import "github.com/google/uuid"

type Actor struct {
	ID    uuid.UUID
	TaxId string
	Type  ActorType
}

func New(taxId string, t ActorType) Actor {
	return Actor{
		ID:    uuid.New(),
		TaxId: taxId,
		Type:  t,
	}
}

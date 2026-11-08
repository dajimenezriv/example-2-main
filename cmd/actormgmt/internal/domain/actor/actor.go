package actor

import (
	"errors"
	"uuid"
)

var (
	ErrActorNotFound      = errors.New("actor not found")
	ErrActorAlreadyExists = errors.New("actor already exists")
)

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

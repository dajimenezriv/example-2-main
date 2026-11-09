package actor

import (
	"errors"
	"uuid"
)

var (
	ErrInvalidActorType   = errors.New("invalid actor type")
	ErrActorNotFound      = errors.New("actor not found")
	ErrActorAlreadyExists = errors.New("actor already exists")
)

type Actor struct {
	ID    uuid.UUID
	TaxId string
	Type  ActorType
}

func New(taxId string, t ActorType) (Actor, error) {
	if !t.IsValid() {
		return Actor{}, ErrInvalidActorType
	}

	return Actor{
		ID:    uuid.New(),
		TaxId: taxId,
		Type:  t,
	}, nil
}

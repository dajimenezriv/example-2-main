package actor

import (
	"errors"
	"uuid"
)

var (
	ErrTaxIDRequired      = errors.New("tax ID required")
	ErrInvalidActorType   = errors.New("invalid actor type")
	ErrActorNotFound      = errors.New("actor not found")
	ErrActorAlreadyExists = errors.New("actor already exists")
)

type Actor struct {
	ID    uuid.UUID
	TaxId string
	Type  ActorType
}

func New(id uuid.UUID, taxID string, t ActorType) (Actor, error) {
	if taxID == "" {
		return Actor{}, ErrTaxIDRequired
	}
	if !t.IsValid() {
		return Actor{}, ErrInvalidActorType
	}

	return Actor{
		ID:    id,
		TaxId: taxID,
		Type:  t,
	}, nil
}

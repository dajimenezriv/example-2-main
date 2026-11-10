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

type ActorType string

const (
	Issuer             ActorType = "issuer"
	Investor           ActorType = "investor"
	PlatformAdmin      ActorType = "platform_admin"
	MarketOperator     ActorType = "market_operator"
	SettlementOperator ActorType = "settlement_operator"
	Custom             ActorType = "custom"
)

var validTypes = map[ActorType]struct{}{
	Issuer:             {},
	Investor:           {},
	PlatformAdmin:      {},
	MarketOperator:     {},
	SettlementOperator: {},
	Custom:             {},
}

func (t ActorType) IsValid() bool {
	_, ok := validTypes[t]
	return ok
}

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

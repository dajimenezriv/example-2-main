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

type Type string

const (
	Issuer             Type = "issuer"
	Investor           Type = "investor"
	PlatformAdmin      Type = "platform_admin"
	MarketOperator     Type = "market_operator"
	SettlementOperator Type = "settlement_operator"
	Custom             Type = "custom"
)

var validTypes = map[Type]struct{}{
	Issuer:             {},
	Investor:           {},
	PlatformAdmin:      {},
	MarketOperator:     {},
	SettlementOperator: {},
	Custom:             {},
}

func (t Type) IsValid() bool {
	_, ok := validTypes[t]
	return ok
}

type Actor struct {
	ID    uuid.UUID
	TaxID string
	Type  Type
}

func New(id uuid.UUID, taxID string, t Type) (Actor, error) {
	if taxID == "" {
		return Actor{}, ErrTaxIDRequired
	}
	if !t.IsValid() {
		return Actor{}, ErrInvalidActorType
	}

	return Actor{
		ID:    id,
		TaxID: taxID,
		Type:  t,
	}, nil
}

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

func (t Type) IsValid() bool {
	switch t {
	case Issuer, Investor, PlatformAdmin, MarketOperator, SettlementOperator, Custom:
		return true
	}
	return false
}

type Actor struct {
	id    uuid.UUID
	taxID string
	typ   Type
}

func New(id uuid.UUID, taxID string, typ Type) (*Actor, error) {
	if taxID == "" {
		return nil, ErrTaxIDRequired
	}
	if !typ.IsValid() {
		return nil, ErrInvalidActorType
	}

	return &Actor{
		id:    id,
		taxID: taxID,
		typ:   typ,
	}, nil
}

// Reconstitute rebuilds an Actor from persistence.
// Intended for repository adapters only.
func Reconstitute(id uuid.UUID, taxID string, typ Type) *Actor {
	return &Actor{
		id:    id,
		taxID: taxID,
		typ:   typ,
	}
}

func (a *Actor) ID() uuid.UUID { return a.id }
func (a *Actor) TaxID() string { return a.taxID }
func (a *Actor) Type() Type    { return a.typ }

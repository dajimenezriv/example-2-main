package actor

import (
	"errors"
	"regexp"
	"uuid"
)

var (
	ErrInvalidTaxID       = errors.New("actor: invalid tax ID")
	ErrInvalidType        = errors.New("actor: invalid type")
	ErrNotFound           = errors.New("actor: not found")
	ErrAlreadyExists      = errors.New("actor: already exists")
	ErrTaxIDAlreadyExists = errors.New("actor: tax ID is taken")
)

type (
	TaxID string
	Type  string
)

var taxIDRegex = regexp.MustCompile(`^[0-9]{9}$`)

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
	taxID TaxID
	typ   Type
}

func New(id uuid.UUID, taxID string, typ Type) (*Actor, error) {
	switch {
	case !taxIDRegex.MatchString(taxID):
		return nil, ErrInvalidTaxID
	case !typ.IsValid():
		return nil, ErrInvalidType
	}

	return &Actor{
		id:    id,
		taxID: TaxID(taxID),
		typ:   typ,
	}, nil
}

// Reconstitute rebuilds an Actor from persistence.
// Intended for repository adapters only.
func Reconstitute(id uuid.UUID, taxID string, typ Type) *Actor {
	return &Actor{
		id:    id,
		taxID: TaxID(taxID),
		typ:   typ,
	}
}

func (a *Actor) ID() uuid.UUID { return a.id }
func (a *Actor) TaxID() TaxID  { return a.taxID }
func (a *Actor) Type() Type    { return a.typ }

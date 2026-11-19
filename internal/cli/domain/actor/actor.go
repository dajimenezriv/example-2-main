package actor

import "uuid"

type Type string

const (
	TypeIssuer             Type = "issuer"
	TypeInvestor           Type = "investor"
	TypePlatformAdmin      Type = "platform_admin"
	TypeMarketOperator     Type = "market_operator"
	TypeSettlementOperator Type = "settlement_operator"
	TypeCustom             Type = "custom"
)

func Types() []Type {
	return []Type{TypeIssuer, TypeInvestor, TypePlatformAdmin, TypeMarketOperator, TypeSettlementOperator, TypeCustom}
}

type Actor struct {
	ID       uuid.UUID
	TaxID    string
	Type     Type
	Version  int64
	Accounts []Account
}

type Account struct {
	ID           uuid.UUID
	DltAccountID string
	PublicKey    string
}

type Page struct {
	Actors      []Actor
	TotalCount  int64
	HasNext     bool
	HasPrevious bool
}

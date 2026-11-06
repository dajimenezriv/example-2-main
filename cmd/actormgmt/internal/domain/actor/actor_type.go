package actor

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

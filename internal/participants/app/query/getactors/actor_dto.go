package getactors

import "fmt"

type ActorDTO struct {
	ID       string       `json:"id"`
	TaxId    string       `json:"tax_id"`
	Type     string       `json:"type"`
	Accounts []AccountDTO `json:"accounts"`
}

type AccountDTO struct {
	ID           string `json:"id"`
	DltAccountId string `json:"dlt_account_id"`
	PublicKey    string `json:"public_key"`
}

func (a ActorDTO) String() string {
	return fmt.Sprintf("ActorDTO{ID: %s, TaxId: %s, Type: %s}", a.ID, a.TaxId, a.Type)
}

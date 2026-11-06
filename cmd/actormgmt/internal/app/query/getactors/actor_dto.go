package getactors

import "fmt"

type ActorDTO struct {
	ID    string `json:"id"`
	TaxId string `json:"tax_id"`
	Type  string `json:"type"`
}

func (a ActorDTO) String() string {
	return fmt.Sprintf("ActorDTO{ID: %s, TaxId: %s, Type: %s}", a.ID, a.TaxId, a.Type)
}

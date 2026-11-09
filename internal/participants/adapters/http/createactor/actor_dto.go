package createactor

type CreateActorRequest struct {
	TaxId string `json:"tax_id"`
	Type  string `json:"type"`
}

type CreateActorResponse struct {
	ID    string `json:"id"`
	TaxId string `json:"tax_id"`
	Type  string `json:"type"`
}

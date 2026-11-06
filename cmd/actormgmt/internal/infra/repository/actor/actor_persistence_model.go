package actorrepo

import "github.com/google/uuid"

type Actor struct {
	ID    uuid.UUID
	TaxId string
	Type  string
}

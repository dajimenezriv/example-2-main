package actorrepo

import "uuid"

type Actor struct {
	ID    uuid.UUID
	TaxId string
	Type  string
}

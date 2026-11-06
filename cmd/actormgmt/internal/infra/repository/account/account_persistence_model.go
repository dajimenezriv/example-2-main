package accountrepo

import "github.com/google/uuid"

type Account struct {
	ID           uuid.UUID
	DltAccountId string
	PrivateKey   string
	PublicKey    string
}

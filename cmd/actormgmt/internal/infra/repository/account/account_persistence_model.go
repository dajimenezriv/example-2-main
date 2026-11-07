package accountrepo

import "github.com/google/uuid"

type Account struct {
	ID           uuid.UUID
	ActorID      uuid.UUID
	DltAccountId string
	PrivateKey   string
	PublicKey    string
}

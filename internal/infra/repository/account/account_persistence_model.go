package accountrepo

import "uuid"

type Account struct {
	ID           uuid.UUID
	ActorID      uuid.UUID
	DltAccountId string
	PrivateKey   string
	PublicKey    string
}

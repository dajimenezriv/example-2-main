package account

import (
	"errors"

	"github.com/google/uuid"
)

var ErrAccountNotFound = errors.New("account not found")

type Account struct {
	ID           uuid.UUID
	ActorID      uuid.UUID
	DltAccountId string
	PrivateKey   string
	PublicKey    string
}

func New(actorID uuid.UUID, dltId, privKey, pubKey string) Account {
	return Account{
		ID:           uuid.New(),
		ActorID:      actorID,
		DltAccountId: dltId,
		PrivateKey:   privKey,
		PublicKey:    pubKey,
	}
}

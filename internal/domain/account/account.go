package account

import (
	"errors"
	"uuid"
)

var (
	ErrAccountNotFound      = errors.New("account not found")
	ErrAccountAlreadyExists = errors.New("account already exists")
)

type Account struct {
	ID           uuid.UUID
	ActorID      uuid.UUID
	DltAccountId string
	PrivateKey   string
	PublicKey    string
}

func New(id, actorID uuid.UUID, dltId, privKey, pubKey string) Account {
	return Account{
		ID:           id,
		ActorID:      actorID,
		DltAccountId: dltId,
		PrivateKey:   privKey,
		PublicKey:    pubKey,
	}
}

package account

import (
	"errors"
	"uuid"
)

var (
	ErrDltAccountIDRequired = errors.New("DLT account ID is required")
	ErrAccountNotFound      = errors.New("account not found")
	ErrAccountAlreadyExists = errors.New("account already exists")
)

type Account struct {
	ID           uuid.UUID
	ActorID      uuid.UUID
	DltAccountID string
	PrivateKey   string
	PublicKey    string
}

func New(id, actorID uuid.UUID, dltID, privKey, pubKey string) (Account, error) {
	if dltID == "" {
		return Account{}, ErrDltAccountIDRequired
	}

	return Account{
		ID:           id,
		ActorID:      actorID,
		DltAccountID: dltID,
		PrivateKey:   privKey,
		PublicKey:    pubKey,
	}, nil
}

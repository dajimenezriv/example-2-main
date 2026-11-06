package account

import "github.com/google/uuid"

type Account struct {
	ID           uuid.UUID
	DltAccountId string
	PrivateKey   string
	PublicKey    string
}

func New(dltId, privKey, pubKey string) Account {
	return Account{
		ID:           uuid.New(),
		DltAccountId: dltId,
		PrivateKey:   privKey,
		PublicKey:    pubKey,
	}
}

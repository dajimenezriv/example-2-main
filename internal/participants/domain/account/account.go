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
	id           uuid.UUID
	actorID      uuid.UUID
	dltAccountID string
	privateKey   string
	publicKey    string
}

func New(id, actorID uuid.UUID, dltID, privKey, pubKey string) (*Account, error) {
	if dltID == "" {
		return nil, ErrDltAccountIDRequired
	}

	return &Account{
		id:           id,
		actorID:      actorID,
		dltAccountID: dltID,
		privateKey:   privKey,
		publicKey:    pubKey,
	}, nil
}

// Reconstitute rebuilds an Account from persistence.
// Intended for repository adapters only.
func Reconstitute(id, actorID uuid.UUID, dltID, privKey, pubKey string) *Account {
	return &Account{
		id:           id,
		actorID:      actorID,
		dltAccountID: dltID,
		privateKey:   privKey,
		publicKey:    pubKey,
	}
}

func (a *Account) ID() uuid.UUID        { return a.id }
func (a *Account) ActorID() uuid.UUID   { return a.actorID }
func (a *Account) DltAccountID() string { return a.dltAccountID }
func (a *Account) PrivateKey() string   { return a.privateKey }
func (a *Account) PublicKey() string    { return a.publicKey }

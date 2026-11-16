package account

import (
	"errors"
	"uuid"
)

var (
	ErrDltAccountIDRequired = errors.New("account: DLT account ID is required")
	ErrPrivateKeyRequired   = errors.New("account: private key is required")
	ErrPublicKeyRequired    = errors.New("account: public key is required")
	ErrNotFound             = errors.New("account: not found")
	ErrAlreadyExists        = errors.New("account: already exists")
)

// MaxPageLimit is the largest page size for listing accounts.
const MaxPageLimit = 100

type Account struct {
	id           uuid.UUID
	actorID      uuid.UUID
	dltAccountID string
	privateKey   string
	publicKey    string
	version      int64
}

func New(id, actorID uuid.UUID, dltID, privKey, pubKey string) (*Account, error) {
	switch {
	case dltID == "":
		return nil, ErrDltAccountIDRequired
	case privKey == "":
		return nil, ErrPrivateKeyRequired
	case pubKey == "":
		return nil, ErrPublicKeyRequired
	}

	return &Account{
		id:           id,
		actorID:      actorID,
		dltAccountID: dltID,
		privateKey:   privKey,
		publicKey:    pubKey,
		version:      1,
	}, nil
}

// Reconstitute rebuilds an Account from persistence.
// Intended for repository adapters only.
func Reconstitute(id, actorID uuid.UUID, dltID, privKey, pubKey string, version int64) *Account {
	return &Account{
		id:           id,
		actorID:      actorID,
		dltAccountID: dltID,
		privateKey:   privKey,
		publicKey:    pubKey,
		version:      version,
	}
}

func (a *Account) ID() uuid.UUID        { return a.id }
func (a *Account) ActorID() uuid.UUID   { return a.actorID }
func (a *Account) DltAccountID() string { return a.dltAccountID }
func (a *Account) PrivateKey() string   { return a.privateKey }
func (a *Account) PublicKey() string    { return a.publicKey }
func (a *Account) Version() int64       { return a.version }

func (a *Account) ChangeDltAccountID(dltID string) error {
	if dltID == "" {
		return ErrDltAccountIDRequired
	}
	a.dltAccountID = dltID
	return nil
}

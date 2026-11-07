package account

import "github.com/google/uuid"

type Repository interface {
	Save(account *Account) error
	FindByID(id uuid.UUID) (*Account, error)
	FindAll() ([]*Account, error)
	FindByActorIDs(actorIDs []uuid.UUID) ([]*Account, error)
}

package actor

import "github.com/google/uuid"

type Repository interface {
	Save(actor *Actor) error
	FindByID(id uuid.UUID) (*Actor, error)
	FindAll() ([]*Actor, error)
}

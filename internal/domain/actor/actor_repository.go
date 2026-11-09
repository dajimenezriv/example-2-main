package actor

import "uuid"

type Repository interface {
	Save(actor *Actor) error
	FindByID(id uuid.UUID) (*Actor, error)
	FindAll() ([]*Actor, error)
}

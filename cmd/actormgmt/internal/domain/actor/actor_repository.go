package actor

import "uuid"

//go:generate mockgen -source=actor_repository.go -destination=mocks/actor_repository_mock.go -package=mocks
type Repository interface {
	Save(actor *Actor) error
	FindByID(id uuid.UUID) (*Actor, error)
	FindAll() ([]*Actor, error)
}

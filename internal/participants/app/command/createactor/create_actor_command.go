package createactor

import (
	"context"
	"uuid"

	"exercise/internal/participants/domain/actor"
)

type Command struct {
	TaxID string
	Type  actor.Type
}

type Handler struct {
	ActorRepository actor.Repository
	newID           func() uuid.UUID
}

func NewHandler(repo actor.Repository, newID func() uuid.UUID) *Handler {
	return &Handler{ActorRepository: repo, newID: newID}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (*actor.Actor, error) {
	actor, err := actor.New(h.newID(), cmd.TaxID, cmd.Type)
	if err != nil {
		return nil, err
	}
	if err := h.ActorRepository.Save(ctx, &actor); err != nil {
		return nil, err
	}
	return &actor, nil
}

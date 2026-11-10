package createactor

import (
	"context"
	"uuid"

	actorDomain "exercise/internal/participants/domain/actor"
)

type Command struct {
	TaxID string
	Type  actorDomain.ActorType
}

type Handler struct {
	ActorRepository actorDomain.Repository
	newID           func() uuid.UUID
}

func NewHandler(repo actorDomain.Repository, newID func() uuid.UUID) *Handler {
	return &Handler{ActorRepository: repo, newID: newID}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (*actorDomain.Actor, error) {
	actor, err := actorDomain.New(h.newID(), cmd.TaxID, cmd.Type)
	if err != nil {
		return nil, err
	}

	if err := h.ActorRepository.Save(ctx, &actor); err != nil {
		return nil, err
	}

	return &actor, nil
}

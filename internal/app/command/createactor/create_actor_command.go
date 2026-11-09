package createactor

import (
	"context"
	"errors"
	"uuid"

	actorDomain "exercise/internal/domain/actor"
)

type Command struct {
	TaxId string
	Type  actorDomain.ActorType
}

type Handler struct {
	ActorRepository actorDomain.Repository
	newID           func() uuid.UUID
}

func NewHandler(repo actorDomain.Repository, newID func() uuid.UUID) *Handler {
	return &Handler{ActorRepository: repo}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (*actorDomain.Actor, error) {
	if cmd.TaxId == "" {
		return nil, errors.New("TaxId is required")
	}

	actor, err := actorDomain.New(h.newID(), cmd.TaxId, cmd.Type)
	if err != nil {
		return nil, err
	}

	if err := h.ActorRepository.Save(ctx, &actor); err != nil {
		return nil, err
	}

	return &actor, nil
}

package createactor

import (
	"context"
	"errors"

	actorDomain "exercise2/cmd/actormgmt/internal/domain/actor"
)

type Command struct {
	TaxId string
	Type  actorDomain.ActorType
}

type Handler struct {
	ActorRepository actorDomain.Repository
}

func NewHandler(repo actorDomain.Repository) *Handler {
	return &Handler{ActorRepository: repo}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (*actorDomain.Actor, error) {
	if cmd.TaxId == "" {
		return nil, errors.New("TaxId is required")
	}

	actor := actorDomain.New(cmd.TaxId, cmd.Type)

	if err := h.ActorRepository.Save(&actor); err != nil {
		return nil, err
	}

	return &actor, nil
}

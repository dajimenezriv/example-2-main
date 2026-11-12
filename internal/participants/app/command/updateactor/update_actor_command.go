package updateactor

import (
	"context"
	"uuid"

	"exercise/internal/participants/domain/actor"
)

type Command struct {
	ID    uuid.UUID
	TaxID string
	Type  actor.Type
}

type Handler struct {
	repo actor.Repository
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (*actor.Actor, error) {
	a, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if err := a.ChangeTaxID(cmd.TaxID); err != nil {
		return nil, err
	}
	if err := a.ChangeType(cmd.Type); err != nil {
		return nil, err
	}
	if err := h.repo.Save(ctx, a); err != nil {
		return nil, err
	}
	return a, nil
}

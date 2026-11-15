package updateactor

import (
	"context"
	"uuid"

	"exercise/internal/participants/domain/actor"
)

type Command struct {
	ID      uuid.UUID
	TaxID   *string
	Type    *actor.Type
	Version int64
}

type Handler struct {
	repo actor.Repository
}

func NewHandler(repo actor.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (*actor.Actor, error) {
	a, err := h.repo.FindByID(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}
	if a.Version() != cmd.Version {
		return nil, actor.ErrVersionConflict
	}
	if cmd.TaxID != nil {
		if err := a.ChangeTaxID(*cmd.TaxID); err != nil {
			return nil, err
		}
	}
	if cmd.Type != nil {
		if err := a.ChangeType(*cmd.Type); err != nil {
			return nil, err
		}
	}
	return h.repo.Update(ctx, a, cmd.Version)
}

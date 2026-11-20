package deleteactor

import (
	"context"
	"fmt"
	"uuid"

	"exercise/internal/participants/domain/actor"
)

type Command struct {
	ID      uuid.UUID
	Version int64
}

type Handler struct {
	repo actor.Repository
}

func NewHandler(repo actor.Repository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) error {
	if err := h.repo.Delete(ctx, cmd.ID, cmd.Version); err != nil {
		return fmt.Errorf("delete actor: %w", err)
	}
	return nil
}

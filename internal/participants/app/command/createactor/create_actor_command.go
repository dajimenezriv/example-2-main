package createactor

import (
	"context"
	"fmt"
	"uuid"

	"exercise/internal/participants/domain/actor"
)

type Command struct {
	TaxID string
	Type  actor.Type
}

type Handler struct {
	repo  actor.Repository
	newID func() uuid.UUID
}

func NewHandler(repo actor.Repository, newID func() uuid.UUID) *Handler {
	return &Handler{repo: repo, newID: newID}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (*actor.Actor, error) {
	a, err := actor.New(h.newID(), cmd.TaxID, cmd.Type)
	if err != nil {
		return nil, fmt.Errorf("build actor: %w", err)
	}
	if err := h.repo.Save(ctx, a); err != nil {
		return nil, fmt.Errorf("save actor: %w", err)
	}
	return a, nil
}

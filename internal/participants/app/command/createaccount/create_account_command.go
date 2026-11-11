package createaccount

import (
	"context"
	"fmt"
	"uuid"

	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/actor"
)

type Command struct {
	ActorID      uuid.UUID
	DltAccountID string
	PrivateKey   string
	PublicKey    string
}

type Handler struct {
	AccountRepository account.Repository
	ActorRepository   actor.Repository
	newID             func() uuid.UUID
}

func NewHandler(accountRepo account.Repository, actorRepo actor.Repository, newID func() uuid.UUID) *Handler {
	return &Handler{
		AccountRepository: accountRepo,
		ActorRepository:   actorRepo,
		newID:             newID,
	}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (*account.Account, error) {
	acc, err := account.New(h.newID(), cmd.ActorID, cmd.DltAccountID, cmd.PrivateKey, cmd.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("build account: %w", err)
	}
	if _, err := h.ActorRepository.FindByID(ctx, cmd.ActorID); err != nil {
		return nil, fmt.Errorf("find actor %s: %w", cmd.ActorID, err)
	}
	if err := h.AccountRepository.Save(ctx, &acc); err != nil {
		return nil, fmt.Errorf("save account: %w", err)
	}
	return &acc, nil
}

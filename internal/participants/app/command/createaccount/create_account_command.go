package createaccount

import (
	"context"
	"uuid"

	accountDomain "exercise/internal/participants/domain/account"
	actorDomain "exercise/internal/participants/domain/actor"
)

type Command struct {
	ActorID      uuid.UUID
	DltAccountId string
	PrivateKey   string
	PublicKey    string
}

type Handler struct {
	AccountRepository accountDomain.Repository
	ActorRepository   actorDomain.Repository
	newID             func() uuid.UUID
}

func NewHandler(account accountDomain.Repository, actor actorDomain.Repository, newID func() uuid.UUID) *Handler {
	return &Handler{AccountRepository: account, ActorRepository: actor, newID: newID}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (*accountDomain.Account, error) {
	acc, err := accountDomain.New(h.newID(), cmd.ActorID, cmd.DltAccountId, cmd.PrivateKey, cmd.PublicKey)
	if err != nil {
		return nil, err
	}
	if _, err := h.ActorRepository.FindByID(ctx, cmd.ActorID); err != nil {
		return nil, err
	}
	if err := h.AccountRepository.Save(ctx, &acc); err != nil {
		return nil, err
	}
	return &acc, nil
}

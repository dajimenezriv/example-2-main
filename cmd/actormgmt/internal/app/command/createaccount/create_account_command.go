package createaccount

import (
	"context"
	"errors"

	accountDomain "exercise2/cmd/actormgmt/internal/domain/account"
	actorDomain "exercise2/cmd/actormgmt/internal/domain/actor"

	"github.com/google/uuid"
)

var ErrDltAccountIdRequired = errors.New("DltAccountId is required")

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
	if cmd.DltAccountId == "" {
		return nil, ErrDltAccountIdRequired
	}

	act, err := h.ActorRepository.FindByID(cmd.ActorID)
	if err != nil {
		return nil, err
	}

	acc := accountDomain.New(h.newID(), act.ID, cmd.DltAccountId, cmd.PrivateKey, cmd.PublicKey)

	if err := h.AccountRepository.Save(&acc); err != nil {
		return nil, err
	}

	return &acc, nil
}

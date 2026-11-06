package createaccount

import (
	"context"
	"errors"

	accountDomain "exercise2/cmd/actormgmt/internal/domain/account"
)

type Command struct {
	DltAccountId string
	PrivateKey   string
	PublicKey    string
}

type Handler struct {
	AccountRepository accountDomain.Repository
}

func NewHandler(repo accountDomain.Repository) *Handler {
	return &Handler{AccountRepository: repo}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (*accountDomain.Account, error) {
	if cmd.DltAccountId == "" {
		return nil, errors.New("DltAccountId is required")
	}

	acc := accountDomain.New(cmd.DltAccountId, cmd.PrivateKey, cmd.PublicKey)

	if err := h.AccountRepository.Save(&acc); err != nil {
		return nil, err
	}

	return &acc, nil
}

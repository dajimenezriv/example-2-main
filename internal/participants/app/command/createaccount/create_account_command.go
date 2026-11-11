package createaccount

import (
	"context"
	"fmt"
	"uuid"

	"exercise/internal/participants/app/uow"
	"exercise/internal/participants/domain/account"
)

type Command struct {
	ActorID      uuid.UUID
	DltAccountID string
	PrivateKey   string
	PublicKey    string
}

type Handler struct {
	UnitOfWork uow.UnitOfWork
	newID      func() uuid.UUID
}

func NewHandler(unitOfWork uow.UnitOfWork, newID func() uuid.UUID) *Handler {
	return &Handler{
		UnitOfWork: unitOfWork,
		newID:      newID,
	}
}

func (h *Handler) Execute(ctx context.Context, cmd Command) (*account.Account, error) {
	acc, buildErr := account.New(h.newID(), cmd.ActorID, cmd.DltAccountID, cmd.PrivateKey, cmd.PublicKey)
	if buildErr != nil {
		return nil, fmt.Errorf("build account: %w", buildErr)
	}

	err := h.UnitOfWork.Do(ctx, func(ctx context.Context, repos uow.Repositories) error {
		if _, err := repos.Actors.FindByID(ctx, cmd.ActorID); err != nil {
			return fmt.Errorf("find actor %s: %w", cmd.ActorID, err)
		}
		if err := repos.Accounts.Save(ctx, &acc); err != nil {
			return fmt.Errorf("save account: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("unit of work: %w", err)
	}
	return &acc, nil
}

package createaccount

import (
	"context"
	"errors"
	"net/http"

	"exercise/internal/participants/app/command/createaccount"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/actor"

	"uuid"

	"github.com/danielgtaylor/huma/v2"
)

type Input struct {
	Body createAccountRequest
}

type createAccountRequest struct {
	ActorID      string `json:"actor_id" format:"uuid" doc:"ID of the actor that owns the account"`
	DltAccountID string `json:"dlt_account_id" minLength:"1" doc:"Account ID on the DLT"`
	PrivateKey   string `json:"private_key,omitempty"`
	PublicKey    string `json:"public_key,omitempty"`
}
type Output struct {
	Body createAccountResponse
}

type createAccountResponse struct {
	ID           string `json:"id" format:"uuid"`
	ActorID      string `json:"actor_id" format:"uuid"`
	DltAccountID string `json:"dlt_account_id"`
	PublicKey    string `json:"public_key"`
}

type Handler struct {
	CreateAccount *createaccount.Handler
}

func NewHandler(handler *createaccount.Handler) *Handler {
	return &Handler{CreateAccount: handler}
}

func (h *Handler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		Method:        http.MethodPost,
		Path:          "/accounts",
		Summary:       "Create an account",
		Tags:          []string{"Accounts"},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusBadRequest, http.StatusNotFound, http.StatusConflict},
	}, h.Handle)
}

func (h *Handler) Handle(ctx context.Context, in *Input) (*Output, error) {
	actorID, err := uuid.Parse(in.Body.ActorID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid actor id")
	}

	acc, err := h.CreateAccount.Execute(ctx, createaccount.Command{
		ActorID:      actorID,
		DltAccountID: in.Body.DltAccountID,
		PrivateKey:   in.Body.PrivateKey,
		PublicKey:    in.Body.PublicKey,
	})
	switch {
	case err == nil:
	case errors.Is(err, actor.ErrActorNotFound):
		return nil, huma.Error404NotFound(err.Error())
	case errors.Is(err, account.ErrAccountAlreadyExists):
		return nil, huma.Error409Conflict(err.Error())
	default:
		return nil, huma.Error500InternalServerError(err.Error())
	}

	return &Output{Body: createAccountResponse{
		ID:           acc.ID.String(),
		ActorID:      acc.ActorID.String(),
		DltAccountID: acc.DltAccountID,
		PublicKey:    acc.PublicKey,
	}}, nil
}

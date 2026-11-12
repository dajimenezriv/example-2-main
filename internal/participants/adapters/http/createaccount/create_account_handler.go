package createaccount

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"exercise/internal/participants/app/command/createaccount"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/platform/httpapi"

	"github.com/danielgtaylor/huma/v2"
)

type Input struct {
	Body createAccountRequest
}

type createAccountRequest struct {
	ActorID      string `json:"actorId" format:"uuid"`
	DltAccountID string `json:"dltAccountId" minLength:"1"`
	PrivateKey   string `json:"privateKey" minLength:"1"`
	PublicKey    string `json:"publicKey" minLength:"1"`
}
type Output struct {
	Body createAccountResponse
}

type createAccountResponse struct {
	ID           string `json:"id" format:"uuid"`
	ActorID      string `json:"actorId" format:"uuid"`
	DltAccountID string `json:"dltAccountId"`
	PublicKey    string `json:"publicKey"`
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
		Errors:        []int{http.StatusBadRequest, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity},
	}, h.Handle)
}

func (h *Handler) Handle(ctx context.Context, in *Input) (*Output, error) {
	actorID, err := uuid.Parse(in.Body.ActorID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid actorId")
	}

	acc, err := h.CreateAccount.Execute(ctx, createaccount.Command{
		ActorID:      actorID,
		DltAccountID: in.Body.DltAccountID,
		PrivateKey:   in.Body.PrivateKey,
		PublicKey:    in.Body.PublicKey,
	})
	switch {
	case err == nil:
	case errors.Is(err, actor.ErrNotFound):
		return nil, huma.Error404NotFound("actor not found")
	case errors.Is(err, account.ErrAlreadyExists):
		return nil, huma.Error409Conflict("account already exists")
	default:
		return nil, httpapi.Error500InternalServerError(err)
	}

	return &Output{Body: createAccountResponse{
		ID:           acc.ID().String(),
		ActorID:      acc.ActorID().String(),
		DltAccountID: acc.DltAccountID(),
		PublicKey:    acc.PublicKey(),
	}}, nil
}

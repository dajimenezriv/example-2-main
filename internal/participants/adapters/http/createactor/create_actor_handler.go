package createactor

import (
	"context"
	"net/http"

	"exercise/internal/participants/app/command/createactor"
	"exercise/internal/participants/domain/actor"

	"github.com/danielgtaylor/huma/v2"
)

type Input struct {
	Body createActorRequest
}

type createActorRequest struct {
	TaxID string `json:"tax_id" minLength:"1"`
	Type  string `json:"type" enum:"issuer,investor,platform_admin,market_operator,settlement_operator,custom"`
}

type Output struct {
	Body createActorResponse
}

type createActorResponse struct {
	ID    string `json:"id" format:"uuid"`
	TaxID string `json:"tax_id"`
	Type  string `json:"type"`
}

type Handler struct {
	CreateActor *createactor.Handler
}

func NewHandler(cmd *createactor.Handler) *Handler {
	return &Handler{CreateActor: cmd}
}

func (h *Handler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		Method:        http.MethodPost,
		Path:          "/actors",
		Summary:       "Create an actor",
		Tags:          []string{"Actors"},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusBadRequest},
	}, h.Handle)
}

func (h *Handler) Handle(ctx context.Context, in *Input) (*Output, error) {
	act, err := h.CreateActor.Execute(ctx, createactor.Command{
		TaxID: in.Body.TaxID,
		Type:  actor.ActorType(in.Body.Type),
	})
	if err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}

	return &Output{Body: createActorResponse{
		ID:    act.ID.String(),
		TaxID: act.TaxID,
		Type:  string(act.Type),
	}}, nil
}

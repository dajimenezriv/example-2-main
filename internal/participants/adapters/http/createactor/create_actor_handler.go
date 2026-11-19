package createactor

import (
	"context"
	"errors"
	"net/http"

	"exercise/internal/participants/app/command/createactor"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/platform/httpapi"

	"github.com/danielgtaylor/huma/v2"
)

type Input struct {
	Body createActorRequest
}

type createActorRequest struct {
	TaxID taxID  `json:"taxId"`
	Type  string `json:"type" enum:"issuer,investor,platform_admin,market_operator,settlement_operator,custom"`
}

type taxID string

func (taxID) Schema(huma.Registry) *huma.Schema {
	return &huma.Schema{
		Type:               huma.TypeString,
		Pattern:            actor.TaxIDPattern,
		PatternDescription: actor.TaxIDPatternDescription,
		Examples:           []any{actor.TaxIDExample},
	}
}

type Output struct {
	Body createActorResponse
}

type createActorResponse struct {
	ID      string `json:"id" format:"uuid"`
	TaxID   string `json:"taxId"`
	Type    string `json:"type"`
	Version int64  `json:"version"`
}

type Handler struct {
	CreateActor *createactor.Handler
}

func NewHandler(cmd *createactor.Handler) *Handler {
	return &Handler{CreateActor: cmd}
}

func (h *Handler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "create-actor",
		Method:        http.MethodPost,
		Path:          "/actors",
		Summary:       "Create an actor",
		Tags:          []string{"Actors"},
		DefaultStatus: http.StatusCreated,
		Errors:        []int{http.StatusBadRequest, http.StatusConflict, http.StatusUnprocessableEntity},
	}, h.Handle)
}

func (h *Handler) Handle(ctx context.Context, in *Input) (*Output, error) {
	a, err := h.CreateActor.Execute(ctx, createactor.Command{
		TaxID: string(in.Body.TaxID),
		Type:  actor.Type(in.Body.Type),
	})
	switch {
	case err == nil:
	case errors.Is(err, actor.ErrTaxIDAlreadyExists):
		return nil, huma.Error409Conflict("actor with this tax ID already exists")
	case errors.Is(err, actor.ErrAlreadyExists):
		return nil, huma.Error409Conflict("actor already exists")
	default:
		return nil, httpapi.Error500InternalServerError(err)
	}

	return &Output{Body: createActorResponse{
		ID:      a.ID().String(),
		TaxID:   string(a.TaxID()),
		Type:    string(a.Type()),
		Version: a.Version(),
	}}, nil
}

package updateactor

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"github.com/danielgtaylor/huma/v2"

	"exercise/internal/participants/app/command/updateactor"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/platform/httpapi"
)

type Input struct {
	ID   string `path:"id" format:"uuid"`
	Body updateActorRequest
}

type updateActorRequest struct {
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
	Body updateActorResponse
}

type updateActorResponse struct {
	ID    string `json:"id" format:"uuid"`
	TaxID string `json:"taxId"`
	Type  string `json:"type"`
}

type Handler struct {
	UpdateActor *updateactor.Handler
}

func NewHandler(cmd *updateactor.Handler) *Handler {
	return &Handler{UpdateActor: cmd}
}

func (h *Handler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		Method:        http.MethodPut,
		Path:          "/actors/{id}",
		Summary:       "Update an actor",
		Tags:          []string{"Actors"},
		DefaultStatus: http.StatusOK,
		Errors:        []int{http.StatusBadRequest, http.StatusNotFound, http.StatusConflict, http.StatusUnprocessableEntity},
	}, h.Handle)
}

func (h *Handler) Handle(ctx context.Context, in *Input) (*Output, error) {
	id, err := uuid.Parse(in.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid actorId")
	}

	a, err := h.UpdateActor.Execute(ctx, updateactor.Command{
		ID:    id,
		TaxID: string(in.Body.TaxID),
		Type:  actor.Type(in.Body.Type),
	})
	switch {
	case err == nil:
	case errors.Is(err, actor.ErrNotFound):
		return nil, huma.Error404NotFound("actor not found")
	case errors.Is(err, actor.ErrVersionConflict):
		return nil, huma.Error409Conflict("actor version conflict")
	default:
		return nil, httpapi.Error500InternalServerError(err)
	}

	return &Output{Body: updateActorResponse{
		ID:    a.ID().String(),
		TaxID: string(a.TaxID()),
		Type:  string(a.Type()),
	}}, nil
}

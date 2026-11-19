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
	TaxID   *taxID  `json:"taxId,omitempty"`
	Type    *string `json:"type,omitempty" enum:"issuer,investor,platform_admin,market_operator,settlement_operator,custom"`
	Version int64   `json:"version" minimum:"1"`
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
	ID      string `json:"id" format:"uuid"`
	TaxID   string `json:"taxId"`
	Type    string `json:"type"`
	Version int64  `json:"version"`
}

type Handler struct {
	UpdateActor *updateactor.Handler
}

func NewHandler(cmd *updateactor.Handler) *Handler {
	return &Handler{UpdateActor: cmd}
}

func (h *Handler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "update-actor",
		Method:        http.MethodPatch,
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
		return nil, huma.Error400BadRequest("invalid id")
	}

	cmd := updateactor.Command{
		ID:      id,
		Version: in.Body.Version,
	}
	if in.Body.TaxID != nil {
		cmd.TaxID = new(string(*in.Body.TaxID))
	}
	if in.Body.Type != nil {
		cmd.Type = new(actor.Type(*in.Body.Type))
	}

	a, err := h.UpdateActor.Execute(ctx, cmd)
	switch {
	case err == nil:
	case errors.Is(err, actor.ErrNotFound):
		return nil, huma.Error404NotFound("actor not found")
	case errors.Is(err, actor.ErrTaxIDAlreadyExists):
		return nil, huma.Error409Conflict("actor with this tax ID already exists")
	case errors.Is(err, actor.ErrVersionConflict):
		return nil, huma.Error409Conflict("actor version conflict")
	default:
		return nil, httpapi.Error500InternalServerError(err)
	}

	return &Output{Body: updateActorResponse{
		ID:      a.ID().String(),
		TaxID:   string(a.TaxID()),
		Type:    string(a.Type()),
		Version: a.Version(),
	}}, nil
}

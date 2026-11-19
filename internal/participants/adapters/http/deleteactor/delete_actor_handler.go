package deleteactor

import (
	"context"
	"errors"
	"net/http"
	"uuid"

	"github.com/danielgtaylor/huma/v2"

	"exercise/internal/participants/app/command/deleteactor"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/platform/httpapi"
)

type Input struct {
	ID string `path:"id" format:"uuid"`
	// Optional so a missing header gets 428 instead of Huma's 422.
	IfMatch string `header:"If-Match" doc:"Strong ETag of the version being deleted, e.g. \"1\". Required."`
}

type Output struct{}

type Handler struct {
	DeleteActor *deleteactor.Handler
}

func NewHandler(cmd *deleteactor.Handler) *Handler {
	return &Handler{DeleteActor: cmd}
}

func (h *Handler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID:   "delete-actor",
		Method:        http.MethodDelete,
		Path:          "/actors/{id}",
		Summary:       "Delete an actor and its accounts",
		Tags:          []string{"Actors"},
		DefaultStatus: http.StatusNoContent,
		Errors: []int{
			http.StatusBadRequest, http.StatusNotFound, http.StatusPreconditionFailed,
			http.StatusUnprocessableEntity, http.StatusPreconditionRequired,
		},
	}, h.Handle)
}

func (h *Handler) Handle(ctx context.Context, in *Input) (*Output, error) {
	id, err := uuid.Parse(in.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid id")
	}

	version, ifMatchErr := httpapi.VersionFromIfMatch(in.IfMatch)
	if ifMatchErr != nil {
		return nil, ifMatchErr
	}

	err = h.DeleteActor.Execute(ctx, deleteactor.Command{ID: id, Version: version})
	switch {
	case err == nil:
		return &Output{}, nil
	case errors.Is(err, actor.ErrNotFound):
		return nil, huma.Error404NotFound("actor not found")
	case errors.Is(err, actor.ErrVersionConflict):
		return nil, huma.Error412PreconditionFailed("actor version conflict")
	default:
		return nil, httpapi.Error500InternalServerError(err)
	}
}

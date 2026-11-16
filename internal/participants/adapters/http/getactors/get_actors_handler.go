package getactors

import (
	"context"
	"errors"
	"net/http"

	"exercise/internal/participants/app/query/getactors"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/platform/httpapi"

	"github.com/danielgtaylor/huma/v2"
)

type Input struct {
	Limit  int32 `query:"limit" default:"20"`
	Offset int32 `query:"offset" default:"0"`
}

type Output struct {
	Body getActorsResponse
}

type getActorsResponse struct {
	Actors      []actorItem `json:"actors"`
	TotalCount  int64       `json:"totalCount"`
	HasNext     bool        `json:"hasNext"`
	HasPrevious bool        `json:"hasPrevious"`
}

type actorItem struct {
	ID       string             `json:"id" format:"uuid"`
	TaxID    string             `json:"taxId"`
	Type     string             `json:"type"`
	Version  int64              `json:"version"`
	Accounts []actorAccountItem `json:"accounts"`
}

type actorAccountItem struct {
	ID           string `json:"id" format:"uuid"`
	DltAccountID string `json:"dltAccountId"`
	PublicKey    string `json:"publicKey"`
}

type Handler struct {
	Query *getactors.Query
}

func NewHandler(query *getactors.Query) *Handler {
	return &Handler{Query: query}
}

func (h *Handler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		Method:  http.MethodGet,
		Path:    "/actors",
		Summary: "List actors with their accounts",
		Tags:    []string{"Actors"},
	}, h.Handle)
}

func (h *Handler) Handle(ctx context.Context, in *Input) (*Output, error) {
	res, err := h.Query.Execute(ctx, in.Limit, in.Offset)
	switch {
	case err == nil:
	case errors.Is(err, actor.ErrInvalidPagination):
		return nil, huma.Error400BadRequest("invalid pagination")
	default:
		return nil, httpapi.Error500InternalServerError(err)
	}

	resp := getActorsResponse{
		Actors:      make([]actorItem, len(res.Actors)),
		TotalCount:  res.TotalCount,
		HasNext:     res.HasNext,
		HasPrevious: res.HasPrevious,
	}
	for i, act := range res.Actors {
		accountItems := make([]actorAccountItem, len(act.Accounts))
		for j, acc := range act.Accounts {
			accountItems[j] = actorAccountItem{
				ID:           acc.ID.String(),
				DltAccountID: acc.DltAccountID,
				PublicKey:    acc.PublicKey,
			}
		}
		resp.Actors[i] = actorItem{
			ID:       act.ID.String(),
			TaxID:    string(act.TaxID),
			Type:     string(act.Type),
			Version:  act.Version,
			Accounts: accountItems,
		}
	}
	return &Output{Body: resp}, nil
}

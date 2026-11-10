package getactors

import (
	"context"
	"net/http"

	"exercise/internal/participants/app/query/getactors"

	"github.com/danielgtaylor/huma/v2"
)

type Output struct {
	Body getActorsResponse
}

type getActorsResponse struct {
	Actors []actorItem `json:"actors"`
}

type actorItem struct {
	ID       string             `json:"id" format:"uuid"`
	TaxID    string             `json:"tax_id"`
	Type     string             `json:"type"`
	Accounts []actorAccountItem `json:"accounts"`
}

type actorAccountItem struct {
	ID           string `json:"id" format:"uuid"`
	DltAccountID string `json:"dlt_account_id"`
	PublicKey    string `json:"public_key"`
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

func (h *Handler) Handle(ctx context.Context, _ *struct{}) (*Output, error) {
	actors, err := h.Query.Execute(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}

	resp := getActorsResponse{Actors: make([]actorItem, len(actors))}
	for i, act := range actors {
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
			TaxID:    act.TaxID,
			Type:     string(act.Type),
			Accounts: accountItems,
		}
	}
	return &Output{Body: resp}, nil
}

package getactors

import (
	"encoding/json"
	"exercise/internal/participants/app/query/getactors"
	"net/http"
)

type getActorsResponse struct {
	Actors []actorItem `json:"actors"`
}

type actorItem struct {
	ID       string        `json:"id"`
	TaxID    string        `json:"tax_id"`
	Type     string        `json:"type"`
	Accounts []accountItem `json:"accounts"`
}

type accountItem struct {
	ID           string `json:"id"`
	DltAccountID string `json:"dlt_account_id"`
	PublicKey    string `json:"public_key"`
}

type Handler struct {
	Query *getactors.Query
}

func NewHandler(query *getactors.Query) *Handler {
	return &Handler{Query: query}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	actors, err := h.Query.Execute(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := getActorsResponse{Actors: make([]actorItem, len(actors))}
	for i, act := range actors {
		accountItems := make([]accountItem, len(act.Accounts))
		for j, acc := range act.Accounts {
			accountItems[j] = accountItem{
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

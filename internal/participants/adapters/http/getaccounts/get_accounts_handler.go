package getaccounts

import (
	"encoding/json"
	"exercise/internal/participants/app/query/getaccounts"
	"net/http"
)

type getAccountsResponse struct {
	Accounts []accountItem `json:"accounts"`
}

type accountItem struct {
	ID           string `json:"id"`
	DltAccountID string `json:"dlt_account_id"`
	PublicKey    string `json:"public_key"`
}

type Handler struct {
	Query *getaccounts.Query
}

func NewHandler(query *getaccounts.Query) *Handler {
	return &Handler{Query: query}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.Query.Execute(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := getAccountsResponse{Accounts: make([]accountItem, len(accounts))}
	for i, a := range accounts {
		resp.Accounts[i] = accountItem{
			ID:           a.ID,
			DltAccountID: a.DltAccountID,
			PublicKey:    a.PublicKey,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

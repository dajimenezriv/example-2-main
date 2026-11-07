package getaccounts

import (
	"encoding/json"
	"exercise2/cmd/actormgmt/internal/app/query/getaccounts"
	"net/http"
)

type Handler struct {
	Query *getaccounts.Query
}

func NewHandler(query *getaccounts.Query) *Handler {
	return &Handler{Query: query}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.Query.Execute()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accounts)
}

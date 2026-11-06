package getactors

import (
	"encoding/json"
	"exercise2/cmd/actormgmt/internal/app/query/getactors"
	"net/http"
)

type Handler struct {
	Query *getactors.Query
}

func NewHandler(query *getactors.Query) *Handler {
	return &Handler{Query: query}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	actors, err := h.Query.Execute()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(actors)
}

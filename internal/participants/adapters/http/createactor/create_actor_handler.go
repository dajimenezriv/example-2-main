package createactor

import (
	"encoding/json"
	"exercise/internal/participants/app/command/createactor"
	"net/http"

	"exercise/internal/participants/domain/actor"
)

type createActorRequest struct {
	TaxID string `json:"tax_id"`
	Type  string `json:"type"`
}

type createActorResponse struct {
	ID    string `json:"id"`
	TaxID string `json:"tax_id"`
	Type  string `json:"type"`
}

type Handler struct {
	CreateActor *createactor.Handler
}

func NewHandler(cmd *createactor.Handler) *Handler {
	return &Handler{CreateActor: cmd}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req createActorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	actorType := actor.ActorType(req.Type)

	actor, err := h.CreateActor.Execute(r.Context(), createactor.Command{
		TaxID: req.TaxID,
		Type:  actorType,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := createActorResponse{
		ID:    actor.ID.String(),
		TaxID: actor.TaxID,
		Type:  string(actor.Type),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

package createactor

import (
	"encoding/json"
	"exercise/internal/participants/app/command/createactor"
	"net/http"

	"exercise/internal/participants/domain/actor"
)

type Handler struct {
	CreateActor *createactor.Handler
}

func NewHandler(cmd *createactor.Handler) *Handler {
	return &Handler{CreateActor: cmd}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req CreateActorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	actorType := actor.ActorType(req.Type)

	actor, err := h.CreateActor.Execute(r.Context(), createactor.Command{
		TaxId: req.TaxId,
		Type:  actorType,
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := CreateActorResponse{
		ID:    actor.ID.String(),
		TaxId: actor.TaxId,
		Type:  string(actor.Type),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

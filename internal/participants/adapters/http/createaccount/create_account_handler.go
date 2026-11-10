package createaccount

import (
	"encoding/json"
	"errors"
	"net/http"

	"exercise/internal/participants/app/command/createaccount"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/actor"

	"uuid"
)

type createAccountRequest struct {
	ActorID      string `json:"actor_id"`
	DltAccountId string `json:"dlt_account_id"`
	PrivateKey   string `json:"private_key"`
	PublicKey    string `json:"public_key"`
}

type createAccountResponse struct {
	ID           string `json:"id"`
	ActorID      string `json:"actor_id"`
	DltAccountId string `json:"dlt_account_id"`
	PublicKey    string `json:"public_key"`
}

type Handler struct {
	CreateAccount *createaccount.Handler
}

func NewHandler(handler *createaccount.Handler) *Handler {
	return &Handler{CreateAccount: handler}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req createAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	actorID, err := uuid.Parse(req.ActorID)
	if err != nil {
		http.Error(w, "invalid actor id", http.StatusBadRequest)
		return
	}

	acc, err := h.CreateAccount.Execute(r.Context(), createaccount.Command{
		ActorID:      actorID,
		DltAccountId: req.DltAccountId,
		PrivateKey:   req.PrivateKey,
		PublicKey:    req.PublicKey,
	})
	if err != nil {
		if errors.Is(err, actor.ErrActorNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, account.ErrAccountAlreadyExists) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := createAccountResponse{
		ID:           acc.ID.String(),
		ActorID:      acc.ActorID.String(),
		DltAccountId: acc.DltAccountId,
		PublicKey:    acc.PublicKey,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

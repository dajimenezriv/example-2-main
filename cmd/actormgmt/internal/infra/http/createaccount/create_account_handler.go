package createaccount

import (
	"encoding/json"
	"net/http"

	"exercise2/cmd/actormgmt/internal/app/command/createaccount"

	"uuid"
)

type Handler struct {
	CreateAccount *createaccount.Handler
}

func NewHandler(handler *createaccount.Handler) *Handler {
	return &Handler{CreateAccount: handler}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	actorID, err := uuid.Parse(req.ActorID)
	if err != nil {
		http.Error(w, "invalid actor id", http.StatusBadRequest)
		return
	}

	account, err := h.CreateAccount.Execute(r.Context(), createaccount.Command{
		ActorID:      actorID,
		DltAccountId: req.DltAccountId,
		PrivateKey:   req.PrivateKey,
		PublicKey:    req.PublicKey,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := CreateAccountResponse{
		ID:           account.ID.String(),
		ActorID:      account.ActorID.String(),
		DltAccountId: account.DltAccountId,
		PublicKey:    account.PublicKey,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

package createaccount

import (
	"encoding/json"
	"net/http"

	"exercise/internal/participants/app/command/createaccount"

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

	resp := createAccountResponse{
		ID:           account.ID.String(),
		ActorID:      account.ActorID.String(),
		DltAccountId: account.DltAccountId,
		PublicKey:    account.PublicKey,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

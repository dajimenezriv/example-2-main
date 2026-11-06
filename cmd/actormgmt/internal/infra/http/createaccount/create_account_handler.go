package createaccount

import (
	"context"
	"encoding/json"
	"net/http"

	"exercise2/cmd/actormgmt/internal/app/command/createaccount"
)

type Handler struct {
	CreateAccount *createaccount.Handler
}

func NewHandler(handler *createaccount.Handler) *Handler {
	return &Handler{CreateAccount: handler}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	account, err := h.CreateAccount.Execute(context.Background(), createaccount.Command{
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
		DltAccountId: account.DltAccountId,
		PrivateKey:   account.PrivateKey,
		PublicKey:    account.PublicKey,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

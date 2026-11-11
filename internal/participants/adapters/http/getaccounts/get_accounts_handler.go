package getaccounts

import (
	"context"
	"net/http"

	"exercise/internal/participants/app/query/getaccounts"

	"github.com/danielgtaylor/huma/v2"
)

type Output struct {
	Body getAccountsResponse
}

type getAccountsResponse struct {
	Accounts []accountItem `json:"accounts"`
}

type accountItem struct {
	ID           string `json:"id" format:"uuid"`
	DltAccountID string `json:"dltAccountId"`
	PublicKey    string `json:"publicKey"`
}

type Handler struct {
	Query *getaccounts.Query
}

func NewHandler(query *getaccounts.Query) *Handler {
	return &Handler{Query: query}
}

func (h *Handler) Register(api huma.API) {
	huma.Register(api, huma.Operation{
		Method:  http.MethodGet,
		Path:    "/accounts",
		Summary: "List accounts",
		Tags:    []string{"Accounts"},
	}, h.Handle)
}

func (h *Handler) Handle(ctx context.Context, _ *struct{}) (*Output, error) {
	accounts, err := h.Query.Execute(ctx)
	if err != nil {
		return nil, huma.Error500InternalServerError(err.Error())
	}

	resp := getAccountsResponse{Accounts: make([]accountItem, len(accounts))}
	for i, a := range accounts {
		resp.Accounts[i] = accountItem{
			ID:           a.ID,
			DltAccountID: a.DltAccountID,
			PublicKey:    a.PublicKey,
		}
	}
	return &Output{Body: resp}, nil
}

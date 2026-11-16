package getaccounts

import (
	"context"
	"errors"
	"net/http"

	"exercise/internal/participants/app/query/getaccounts"
	"exercise/internal/participants/domain/account"

	"github.com/danielgtaylor/huma/v2"
)

type Input struct {
	Limit  int32 `query:"limit" default:"20"`
	Offset int32 `query:"offset" default:"0"`
}

type Output struct {
	Body getAccountsResponse
}

type getAccountsResponse struct {
	Accounts    []accountItem `json:"accounts"`
	TotalCount  int64         `json:"totalCount"`
	HasNext     bool          `json:"hasNext"`
	HasPrevious bool          `json:"hasPrevious"`
}

type accountItem struct {
	ID           string `json:"id" format:"uuid"`
	DltAccountID string `json:"dltAccountId"`
	PublicKey    string `json:"publicKey"`
	Version      int64  `json:"version"`
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

func (h *Handler) Handle(ctx context.Context, in *Input) (*Output, error) {
	res, err := h.Query.Execute(ctx, in.Limit, in.Offset)
	switch {
	case err == nil:
	case errors.Is(err, account.ErrInvalidPagination):
		return nil, huma.Error400BadRequest("invalid pagination")
	default:
		return nil, huma.Error500InternalServerError(err.Error())
	}

	resp := getAccountsResponse{
		Accounts:    make([]accountItem, len(res.Accounts)),
		TotalCount:  res.TotalCount,
		HasNext:     res.HasNext,
		HasPrevious: res.HasPrevious,
	}
	for i, a := range res.Accounts {
		resp.Accounts[i] = accountItem{
			ID:           a.ID,
			DltAccountID: a.DltAccountID,
			PublicKey:    a.PublicKey,
			Version:      a.Version,
		}
	}
	return &Output{Body: resp}, nil
}

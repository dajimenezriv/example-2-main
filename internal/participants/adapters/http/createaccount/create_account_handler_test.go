package createaccount

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"uuid"

	"exercise/internal/participants/app/command/createaccount"
	"exercise/internal/participants/app/uow"
	"exercise/internal/participants/app/uow/uowtest"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/actor/actortest"
	"exercise/internal/testhelpers"

	"github.com/danielgtaylor/huma/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	actorID   = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	accountID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
)

func newSUT(t *testing.T, newAccountID uuid.UUID) (*Handler, account.Repository) {
	t.Helper()

	act, err := actor.New(actorID, "123456789", actor.Issuer)
	require.NoError(t, err)
	acc, err := account.New(accountID, actorID, "dltID", "privKey", "pubKey")
	require.NoError(t, err)

	actorRepo := actortest.NewFakeRepository()
	accountRepo := accounttest.NewFakeRepository()
	err = actorRepo.Save(t.Context(), act)
	require.NoError(t, err)
	err = accountRepo.Save(t.Context(), acc)
	require.NoError(t, err)

	h := NewHandler(createaccount.NewHandler(
		uowtest.NewFake(uow.Repositories{Accounts: accountRepo, Actors: actorRepo}),
		func() uuid.UUID { return newAccountID },
	))
	return h, accountRepo
}

func requestBody(actorID string) string {
	return fmt.Sprintf(`{
		"actorId": %q,
		"dltAccountId": "dltID",
		"privateKey": "privKey",
		"publicKey": "pubKey"
	}`, actorID)
}

func post(t *testing.T, h *Handler, body string) *httptest.ResponseRecorder {
	t.Helper()
	api := testhelpers.NewHumaAPI(t)
	h.Register(api)
	return api.Post("/accounts", strings.NewReader(body))
}

func TestRejectedRequests(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		newAccountID uuid.UUID
		wantCode     int
		wantDetail   string
		wantErrors   []*huma.ErrorDetail
	}{
		{
			name:         "malformed body",
			body:         ``,
			newAccountID: uuid.New(),
			wantCode:     http.StatusBadRequest,
			wantDetail:   "request body is required",
		},
		{
			name:         "empty body",
			body:         `{}`,
			newAccountID: uuid.New(),
			wantCode:     http.StatusUnprocessableEntity,
			wantDetail:   "validation failed",
			wantErrors: []*huma.ErrorDetail{
				{Message: "expected required property actorId to be present", Location: "body", Value: map[string]any{}},
				{Message: "expected required property dltAccountId to be present", Location: "body", Value: map[string]any{}},
				{Message: "expected required property privateKey to be present", Location: "body", Value: map[string]any{}},
				{Message: "expected required property publicKey to be present", Location: "body", Value: map[string]any{}},
			},
		},
		{
			// Huma accepts an uppercase "URN:UUID:" prefix, but uuid.Parse does not.
			name:         "actor ID rejected by uuid.Parse",
			body:         requestBody("URN:UUID:" + uuid.New().String()),
			newAccountID: uuid.New(),
			wantCode:     http.StatusBadRequest,
			wantDetail:   "invalid actorId",
		},
		{
			name:         "actor not found",
			body:         requestBody(uuid.New().String()),
			newAccountID: uuid.New(),
			wantCode:     http.StatusNotFound,
			wantDetail:   "actor not found",
		},
		{
			name:         "account already exists",
			body:         requestBody(actorID.String()),
			newAccountID: accountID,
			wantCode:     http.StatusConflict,
			wantDetail:   "account already exists",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, accountRepo := newSUT(t, tt.newAccountID)
			rec := post(t, h, tt.body)

			testhelpers.AssertHumaError(t, rec, tt.wantCode, tt.wantDetail, tt.wantErrors...)

			accounts, err := accountRepo.FindAll(t.Context(), accounttest.MustPage(t, 100, 0))
			require.NoError(t, err)
			assert.Len(t, accounts, 1)
		})
	}
}

func TestCreatesAccount(t *testing.T) {
	body := requestBody(actorID.String())

	newAccountID := uuid.New()
	h, accountRepo := newSUT(t, newAccountID)
	rec := post(t, h, body)

	want := fmt.Sprintf(`{
		"id": %q,
		"actorId": %q,
		"dltAccountId": "dltID",
		"publicKey": "pubKey"
	}`, newAccountID, actorID)

	testhelpers.AssertHumaOutput(t, rec, http.StatusCreated, want)

	accounts, err := accountRepo.FindAll(t.Context(), accounttest.MustPage(t, 100, 0))
	require.NoError(t, err)
	assert.Len(t, accounts, 2)
}

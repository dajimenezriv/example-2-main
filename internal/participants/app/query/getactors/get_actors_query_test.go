package getactors

import (
	"context"
	"errors"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/actor/actortest"
	"testing"
	"uuid"

	"github.com/google/go-cmp/cmp"
)

var (
	actorID1   = uuid.New()
	actorID2   = uuid.New()
	actorID3   = uuid.New()
	accountID1 = uuid.New()
	accountID2 = uuid.New()
	accountID3 = uuid.New()
)

func newSUT(t *testing.T) (actor.Repository, account.Repository) {
	t.Helper()
	actorRepo := actortest.NewFakeRepository()
	saveActor(t, actorRepo, actorID1)
	saveActor(t, actorRepo, actorID2)
	saveActor(t, actorRepo, actorID3)
	accountRepo := accounttest.NewFakeRepository()
	saveAccount(t, accountRepo, accountID1, actorID1)
	saveAccount(t, accountRepo, accountID2, actorID1)
	saveAccount(t, accountRepo, accountID3, actorID2)
	return actorRepo, accountRepo
}

func saveActor(t *testing.T, repo actor.Repository, uuid uuid.UUID) *actor.Actor {
	t.Helper()
	a, err := actor.New(uuid, "Tax ID", actor.Issuer)
	if err != nil {
		t.Fatalf("new actor = %v", err)
	}
	if err := repo.Save(context.Background(), &a); err != nil {
		t.Fatalf("actor repo save = %v", err)
	}
	return &a
}

func saveAccount(t *testing.T, repo account.Repository, uuid, actorID uuid.UUID) *account.Account {
	t.Helper()
	a := account.New(uuid, actorID, "DLT account ID", "Private key", "Public key")
	if err := repo.Save(context.Background(), &a); err != nil {
		t.Fatalf("actor repo save = %v", err)
	}
	return &a
}

func TestContextCanceledReturnsError(t *testing.T) {
	actorRepo, accountRepo := newSUT(t)
	q := NewQuery(actorRepo, accountRepo)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	got, err := q.Execute(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Execute() = %v, want %v", err, context.Canceled)
	}
	if got != nil {
		t.Errorf("Execute() = %v, want nil", got)
	}
}

func TestReturnActorsWithAccounts(t *testing.T) {
	actorRepo, accountRepo := newSUT(t)
	q := NewQuery(actorRepo, accountRepo)

	got, err := q.Execute(t.Context())
	if err != nil {
		t.Errorf("Execute() = %v, want nil", err)
	}

	want := []ActorDTO{
		{ID: actorID1, TaxId: "Tax ID", Type: actor.Issuer, Accounts: []AccountDTO{
			{ID: accountID1, DltAccountId: "DLT account ID", PublicKey: "Public key"},
			{ID: accountID2, DltAccountId: "DLT account ID", PublicKey: "Public key"},
		}},
		{ID: actorID2, TaxId: "Tax ID", Type: actor.Issuer, Accounts: []AccountDTO{
			{ID: accountID3, DltAccountId: "DLT account ID", PublicKey: "Public key"},
		}},
		{ID: actorID3, TaxId: "Tax ID", Type: actor.Issuer, Accounts: nil},
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("actors (-want +got):\n%s", diff)
	}
}

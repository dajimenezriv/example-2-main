package accountrepo

import (
	"exercise/internal/domain/account"
	"exercise/postgres"
)

func ToPersistenceModel(a *account.Account) postgres.CreateAccountParams {
	return postgres.CreateAccountParams{
		ID:           a.ID,
		ActorID:      a.ActorID,
		DltAccountID: a.DltAccountId,
		PrivateKey:   a.PrivateKey,
		PublicKey:    a.PublicKey,
	}
}

func ToDomain(d postgres.Account) *account.Account {
	return &account.Account{
		ID:           d.ID,
		ActorID:      d.ActorID,
		DltAccountId: d.DltAccountID,
		PrivateKey:   d.PrivateKey,
		PublicKey:    d.PublicKey,
	}
}

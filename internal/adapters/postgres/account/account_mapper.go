package accountrepo

import (
	"exercise/internal/adapters/postgres/sqlcgen"
	"exercise/internal/domain/account"
)

func ToPersistenceModel(a *account.Account) sqlcgen.CreateAccountParams {
	return sqlcgen.CreateAccountParams{
		ID:           a.ID,
		ActorID:      a.ActorID,
		DltAccountID: a.DltAccountId,
		PrivateKey:   a.PrivateKey,
		PublicKey:    a.PublicKey,
	}
}

func ToDomain(d sqlcgen.Account) *account.Account {
	return &account.Account{
		ID:           d.ID,
		ActorID:      d.ActorID,
		DltAccountId: d.DltAccountID,
		PrivateKey:   d.PrivateKey,
		PublicKey:    d.PublicKey,
	}
}

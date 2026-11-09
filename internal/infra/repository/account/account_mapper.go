package accountrepo

import "exercise/internal/domain/account"

func ToPersistenceModel(a *account.Account) *Account {
	return &Account{
		ID:           a.ID,
		ActorID:      a.ActorID,
		DltAccountId: a.DltAccountId,
		PrivateKey:   a.PrivateKey,
		PublicKey:    a.PublicKey,
	}
}

func ToDomain(d *Account) *account.Account {
	return &account.Account{
		ID:           d.ID,
		ActorID:      d.ActorID,
		DltAccountId: d.DltAccountId,
		PrivateKey:   d.PrivateKey,
		PublicKey:    d.PublicKey,
	}
}

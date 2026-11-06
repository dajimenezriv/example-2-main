package accountrepo

import "exercise2/cmd/actormgmt/internal/domain/account"

func ToPersistenceModel(a *account.Account) *Account {
	return &Account{
		ID:           a.ID,
		DltAccountId: a.DltAccountId,
		PrivateKey:   a.PrivateKey,
		PublicKey:    a.PublicKey,
	}
}

func ToDomain(d *Account) *account.Account {
	return &account.Account{
		ID:           d.ID,
		DltAccountId: d.DltAccountId,
		PrivateKey:   d.PrivateKey,
		PublicKey:    d.PublicKey,
	}
}

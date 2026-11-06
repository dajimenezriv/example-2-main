package accountrepo

import (
	"database/sql"

	"exercise2/cmd/actormgmt/internal/domain/account"

	"github.com/google/uuid"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(a *account.Account) error {
	model := ToPersistenceModel(a)
	_, err := r.db.Exec("INSERT INTO accounts (id, dlt_account_id, private_key, public_key) VALUES ($1,$2,$3,$4)",
		model.ID, model.DltAccountId, model.PrivateKey, model.PublicKey)
	return err
}

func (r *PostgresRepository) FindByID(id uuid.UUID) (*account.Account, error) {
	row := r.db.QueryRow("SELECT id, dlt_account_id, private_key, public_key FROM accounts WHERE id=$1", id)
	var model Account
	var idStr string
	if err := row.Scan(&idStr, &model.DltAccountId, &model.PrivateKey, &model.PublicKey); err != nil {
		return nil, err
	}
	model.ID, _ = uuid.Parse(idStr)
	return ToDomain(&model), nil
}

func (r *PostgresRepository) FindAll() ([]*account.Account, error) {
	rows, err := r.db.Query("SELECT id, dlt_account_id, private_key, public_key FROM accounts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*account.Account
	for rows.Next() {
		var model Account
		var idStr string
		if err := rows.Scan(&idStr, &model.DltAccountId, &model.PrivateKey, &model.PublicKey); err != nil {
			return nil, err
		}
		model.ID, _ = uuid.Parse(idStr)
		accounts = append(accounts, ToDomain(&model))
	}
	return accounts, nil
}

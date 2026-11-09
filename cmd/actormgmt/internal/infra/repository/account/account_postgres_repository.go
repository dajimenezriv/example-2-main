package accountrepo

import (
	"database/sql"
	"errors"
	"uuid"

	"exercise2/cmd/actormgmt/internal/domain/account"

	"github.com/lib/pq"
)

const uniqueViolation = "23505"

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(a *account.Account) error {
	model := ToPersistenceModel(a)
	_, err := r.db.Exec("INSERT INTO accounts (id, actor_id, dlt_account_id, private_key, public_key) VALUES ($1,$2,$3,$4,$5)",
		model.ID, model.ActorID, model.DltAccountId, model.PrivateKey, model.PublicKey)
	if pgErr, ok := errors.AsType[*pq.Error](err); ok && pgErr.Code == uniqueViolation {
		return account.ErrAccountAlreadyExists
	}
	return err
}

func (r *PostgresRepository) FindByID(id uuid.UUID) (*account.Account, error) {
	row := r.db.QueryRow("SELECT id, actor_id, dlt_account_id, private_key, public_key FROM accounts WHERE id=$1", id)
	var model Account
	var idStr string
	var actorIDStr string
	if err := row.Scan(&idStr, &actorIDStr, &model.DltAccountId, &model.PrivateKey, &model.PublicKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, account.ErrAccountNotFound
		}
		return nil, err
	}
	model.ID, _ = uuid.Parse(idStr)
	model.ActorID, _ = uuid.Parse(actorIDStr)
	return ToDomain(&model), nil
}

func (r *PostgresRepository) FindAll() ([]*account.Account, error) {
	rows, err := r.db.Query("SELECT id, actor_id, dlt_account_id, private_key, public_key FROM accounts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*account.Account
	for rows.Next() {
		var model Account
		var idStr string
		var actorIDStr string
		if err := rows.Scan(&idStr, &actorIDStr, &model.DltAccountId, &model.PrivateKey, &model.PublicKey); err != nil {
			return nil, err
		}
		model.ID, _ = uuid.Parse(idStr)
		model.ActorID, _ = uuid.Parse(actorIDStr)
		accounts = append(accounts, ToDomain(&model))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *PostgresRepository) FindByActorIDs(actorIDs []uuid.UUID) ([]*account.Account, error) {
	ids := make([]string, len(actorIDs))
	for i, id := range actorIDs {
		ids[i] = id.String()
	}

	rows, err := r.db.Query(
		"SELECT id, actor_id, dlt_account_id, private_key, public_key FROM accounts WHERE actor_id = ANY($1::uuid[])",
		pq.Array(ids),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*account.Account
	for rows.Next() {
		var model Account
		var idStr string
		var actorIDStr string
		if err := rows.Scan(&idStr, &actorIDStr, &model.DltAccountId, &model.PrivateKey, &model.PublicKey); err != nil {
			return nil, err
		}
		model.ID, _ = uuid.Parse(idStr)
		model.ActorID, _ = uuid.Parse(actorIDStr)
		accounts = append(accounts, ToDomain(&model))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return accounts, nil
}

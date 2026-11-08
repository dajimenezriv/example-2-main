package actorrepo

import (
	"database/sql"
	"errors"

	"exercise2/cmd/actormgmt/internal/domain/actor"

	"uuid"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Save(a *actor.Actor) error {
	model := ToPersistenceModel(a)
	_, err := r.db.Exec("INSERT INTO actors (id, tax_id, type) VALUES ($1,$2,$3)", model.ID, model.TaxId, model.Type)
	return err
}

func (r *PostgresRepository) FindByID(id uuid.UUID) (*actor.Actor, error) {
	row := r.db.QueryRow("SELECT id, tax_id, type FROM actors WHERE id=$1", id)
	var model Actor
	var idStr string
	if err := row.Scan(&idStr, &model.TaxId, &model.Type); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, actor.ErrActorNotFound
		}
		return nil, err
	}
	model.ID, _ = uuid.Parse(idStr)
	return ToDomain(&model), nil
}

func (r *PostgresRepository) FindAll() ([]*actor.Actor, error) {
	rows, err := r.db.Query("SELECT id, tax_id, type FROM actors")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var actors []*actor.Actor
	for rows.Next() {
		var model Actor
		var idStr string
		if err := rows.Scan(&idStr, &model.TaxId, &model.Type); err != nil {
			return nil, err
		}
		model.ID, _ = uuid.Parse(idStr)
		actors = append(actors, ToDomain(&model))
	}
	return actors, nil
}

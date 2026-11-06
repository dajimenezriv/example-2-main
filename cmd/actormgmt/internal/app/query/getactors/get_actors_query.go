package getactors

import (
	domainActor "exercise2/cmd/actormgmt/internal/domain/actor"
)

type Query struct {
	Repo domainActor.Repository
}

func NewQuery(repo domainActor.Repository) *Query {
	return &Query{Repo: repo}
}

func (q *Query) Execute() ([]*ActorDTO, error) {
	actors, err := q.Repo.FindAll()
	if err != nil {
		return nil, err
	}

	var dtos []*ActorDTO
	for _, a := range actors {
		dtos = append(dtos, &ActorDTO{
			ID:    a.ID.String(),
			TaxId: a.TaxId,
			Type:  string(a.Type),
		})
	}
	return dtos, nil
}

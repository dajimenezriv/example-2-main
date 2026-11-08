package actortest

import (
	"exercise2/cmd/actormgmt/internal/domain/actor"
	"testing"
)

func TestFakeRepositoryContract(t *testing.T) {
	RunRepositoryContract(t, func(t *testing.T) actor.Repository {
		t.Helper()
		return NewFakeRepository()
	})
}

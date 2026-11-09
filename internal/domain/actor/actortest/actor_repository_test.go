package actortest

import (
	"exercise2/internal/domain/actor"
	"testing"
)

func TestFakeRepositoryContract(t *testing.T) {
	RunRepositoryContract(t, func(t *testing.T) actor.Repository {
		t.Helper()
		return NewFakeRepository()
	})
}

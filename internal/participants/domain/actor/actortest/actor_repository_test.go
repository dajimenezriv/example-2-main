package actortest

import (
	"exercise/internal/participants/domain/actor"
	"testing"
)

func TestFakeRepositoryContract(t *testing.T) {
	RunRepositoryContract(t, func(t *testing.T) actor.Repository {
		t.Helper()
		return NewFakeRepository()
	})
}

package actortest

import (
	"testing"

	"exercise/internal/participants/domain/actor"
)

func TestFakeRepositoryContract(t *testing.T) {
	RunRepositoryContract(t, func(t *testing.T) actor.Repository {
		t.Helper()
		return NewFakeRepository()
	})
}

package actortest

import (
	"exercise2/cmd/actormgmt/internal/domain/actor"
	"testing"
)

func TestFakeRepositoryContract(t *testing.T) {
	RunRepositoryContract(t, func(_ *testing.T) actor.Repository { return NewFakeRepository() })
}

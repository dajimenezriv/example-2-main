package accounttest

import (
	"testing"
	"uuid"
)

func TestFakeRepositoryContract(t *testing.T) {
	RunRepositoryContract(t, func(t *testing.T) Fixture {
		t.Helper()
		return Fixture{
			Repo: NewFakeRepository(),
			NewActorID: func(t *testing.T) uuid.UUID {
				t.Helper()
				return uuid.New()
			},
		}
	})
}

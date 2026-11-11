package db

import (
	"testing"

	"exercise/internal/participants/adapters/db/sqlcgen"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/actor/actortest"
)

func TestActorRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test")
	}

	q := sqlcgen.New(testPool)

	actortest.RunRepositoryContract(t, func(t *testing.T) actor.Repository {
		t.Helper()
		if _, err := testPool.Exec(t.Context(), "TRUNCATE actors CASCADE"); err != nil {
			t.Fatalf("TRUNCATE actors: %v", err)
		}
		return NewActorRepository(q)
	})
}

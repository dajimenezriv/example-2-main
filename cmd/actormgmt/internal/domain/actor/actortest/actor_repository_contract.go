package actortest

import (
	"exercise2/cmd/actormgmt/internal/domain/actor"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func RunRepositoryContract(t *testing.T, newRepo func(t *testing.T) actor.Repository) {
	t.Run("Save stores the actor", func(t *testing.T) {
		repo := newRepo(t)
		want := actor.New("TaxID", actor.Issuer)

		if err := repo.Save(&want); err != nil {
			t.Fatalf("Save() = %v, want nil", err)
		}

		got, err := repo.FindByID(want.ID)
		if err != nil {
			t.Errorf("FindByID() = %v, want nil", err)
		}
		if diff := cmp.Diff(&want, got); diff != "" {
			t.Errorf("actor mismatch (-want +got):\n%s", diff)
		}
	})
}

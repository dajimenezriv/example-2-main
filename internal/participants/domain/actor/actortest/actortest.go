package actortest

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"exercise/internal/participants/domain/actor"
)

func AssertActor[T *actor.Actor | []*actor.Actor](t *testing.T, got, want T) {
	t.Helper()
	if diff := cmp.Diff(want, got, cmp.AllowUnexported(actor.Actor{})); diff != "" {
		t.Errorf("actor (-want +got):\n%s", diff)
	}
}

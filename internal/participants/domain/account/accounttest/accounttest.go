package accounttest

import (
	"testing"

	"exercise/internal/participants/domain/account"

	"github.com/google/go-cmp/cmp"
)

func AssertAccount[T *account.Account | []*account.Account](t *testing.T, got, want T) {
	t.Helper()
	if diff := cmp.Diff(want, got, cmp.AllowUnexported(account.Account{})); diff != "" {
		t.Errorf("account (-want +got):\n%s", diff)
	}
}

package accounttest

import (
	"exercise2/cmd/actormgmt/internal/domain/account"
	"testing"
)

func TestFakeRepositoryContract(t *testing.T) {
	RunRepositoryContract(t, func(t *testing.T) account.Repository {
		t.Helper()
		return NewFakeRepository()
	})
}

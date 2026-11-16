package account_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"exercise/internal/participants/domain/account"
)

func TestNewPage(t *testing.T) {
	t.Run("rejects out-of-range values", func(t *testing.T) {
		tests := []struct {
			name          string
			limit, offset int32
		}{
			{"zero limit", 0, 0},
			{"negative limit", -1, 0},
			{"limit above max", account.MaxPageLimit + 1, 0},
			{"negative offset", 20, -1},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := account.NewPage(tt.limit, tt.offset)
				require.ErrorIs(t, err, account.ErrInvalidPagination)
			})
		}
	})

	t.Run("accepts boundary values", func(t *testing.T) {
		tests := []struct {
			name          string
			limit, offset int32
		}{
			{"min limit, zero offset", account.MinPageLimit, 0},
			{"max limit, max offset", account.MaxPageLimit, math.MaxInt32},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				p, err := account.NewPage(tt.limit, tt.offset)
				require.NoError(t, err)
				assert.Equal(t, tt.limit, p.Limit())
				assert.Equal(t, tt.offset, p.Offset())
			})
		}
	})
}

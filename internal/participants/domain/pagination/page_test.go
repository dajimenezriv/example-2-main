package pagination_test

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"exercise/internal/participants/domain/pagination"
)

const testMaxLimit = 50

func mustPage(t *testing.T, limit, offset, maxLimit int32) pagination.Page {
	t.Helper()
	p, err := pagination.NewPage(limit, offset, maxLimit)
	require.NoError(t, err)
	return p
}

func TestNewPage(t *testing.T) {
	t.Run("rejects out-of-range values", func(t *testing.T) {
		tests := []struct {
			name                    string
			limit, offset, maxLimit int32
		}{
			{"zero limit", 0, 0, testMaxLimit},
			{"negative limit", -1, 0, testMaxLimit},
			{"min int32 limit", math.MinInt32, 0, testMaxLimit},
			{"limit above max", testMaxLimit + 1, 0, testMaxLimit},
			{"max int32 limit", math.MaxInt32, 0, testMaxLimit},
			{"negative offset", 20, -1, testMaxLimit},
			{"min int32 offset", 20, math.MinInt32, testMaxLimit},
			{"both invalid", 0, -1, testMaxLimit},
			{"max below min limit", pagination.MinLimit, 0, pagination.MinLimit - 1},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, err := pagination.NewPage(tt.limit, tt.offset, tt.maxLimit)
				require.ErrorIs(t, err, pagination.ErrInvalid)
				assert.Zero(t, got)
			})
		}
	})

	t.Run("accepts boundary values", func(t *testing.T) {
		tests := []struct {
			name                    string
			limit, offset, maxLimit int32
		}{
			{"min limit, zero offset", pagination.MinLimit, 0, testMaxLimit},
			{"max limit, zero offset", testMaxLimit, 0, testMaxLimit},
			{"min limit, max offset", pagination.MinLimit, math.MaxInt32, testMaxLimit},
			{"max limit, max offset", testMaxLimit, math.MaxInt32, testMaxLimit},
			{"max equals min limit", pagination.MinLimit, 0, pagination.MinLimit},
			{"max int32 limit allowed by max", math.MaxInt32, 0, math.MaxInt32},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				p := mustPage(t, tt.limit, tt.offset, tt.maxLimit)
				assert.Equal(t, tt.limit, p.Limit())
				assert.Equal(t, tt.offset, p.Offset())
			})
		}
	})
}

func TestPage_HasPrevious(t *testing.T) {
	tests := []struct {
		name   string
		offset int32
		want   bool
	}{
		{"first page", 0, false},
		{"offset one", 1, true},
		{"offset not aligned to limit", 7, true},
		{"max offset", math.MaxInt32, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := mustPage(t, 10, tt.offset, testMaxLimit)
			assert.Equal(t, tt.want, p.HasPrevious())
		})
	}
}

func TestPage_HasNext(t *testing.T) {
	tests := []struct {
		name          string
		limit, offset int32
		total         int64
		want          bool
	}{
		{"empty collection", 10, 0, 0, false},
		{"page smaller than limit", 10, 0, 5, false},
		{"page ends exactly at total", 10, 0, 10, false},
		{"one item after page", 10, 0, 11, true},
		{"middle page", 10, 10, 30, true},
		{"last page ends exactly at total", 10, 20, 30, false},
		{"offset at total", 10, 30, 30, false},
		{"offset past total", 10, 100, 30, false},
		{"offset not aligned to limit", 10, 5, 16, true},
		{"max offset does not overflow", math.MaxInt32, math.MaxInt32, 5, false},
		{"max offset with larger total", math.MaxInt32, math.MaxInt32, 2*math.MaxInt32 + 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := mustPage(t, tt.limit, tt.offset, math.MaxInt32)
			assert.Equal(t, tt.want, p.HasNext(tt.total))
		})
	}
}

package httpapi

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestETag(t *testing.T) {
	assert.Equal(t, `"3"`, ETag(3))
}

func TestVersionFromIfMatch(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		want     int64
		wantCode int
	}{
		{"strong etag", `"3"`, 3, 0},
		{"surrounding spaces", ` "3" `, 3, 0},
		{"missing", ``, 0, http.StatusPreconditionRequired},
		{"blank", `  `, 0, http.StatusPreconditionRequired},
		{"unquoted", `3`, 0, http.StatusBadRequest},
		{"weak etag", `W/"3"`, 0, http.StatusBadRequest},
		{"wildcard", `*`, 0, http.StatusBadRequest},
		{"list", `"1", "2"`, 0, http.StatusBadRequest},
		{"backtick quoted", "`3`", 0, http.StatusBadRequest},
		{"not a number", `"abc"`, 0, http.StatusBadRequest},
		{"zero", `"0"`, 0, http.StatusBadRequest},
		{"negative", `"-1"`, 0, http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := VersionFromIfMatch(tt.header)
			if tt.wantCode == 0 {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
				return
			}
			require.Error(t, err)
			assert.Equal(t, tt.wantCode, err.GetStatus())
		})
	}
}

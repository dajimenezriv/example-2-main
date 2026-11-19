package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

func ETag(version int64) string {
	return strconv.Quote(strconv.FormatInt(version, 10))
}

func VersionFromIfMatch(header string) (int64, huma.StatusError) {
	// Only a single strong ETag: weak tags, lists and "*" can't name the one version the write applies to.
	header = strings.TrimSpace(header)
	if header == "" {
		return 0, huma.NewError(http.StatusPreconditionRequired, "If-Match header is required")
	}
	unquoted, err := strconv.Unquote(header)
	if err != nil || !strings.HasPrefix(header, `"`) {
		return 0, huma.Error400BadRequest(`If-Match must be a single strong ETag, e.g. "1"`)
	}
	v, err := strconv.ParseInt(unquoted, 10, 64)
	if err != nil || v < 1 {
		return 0, huma.Error400BadRequest(`If-Match must be a single strong ETag, e.g. "1"`)
	}
	return v, nil
}

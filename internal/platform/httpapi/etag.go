package httpapi

import (
	"strconv"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

func ETag(version int64) string {
	return strconv.Quote(strconv.FormatInt(version, 10))
}

func VersionFromIfMatch(header string) (int64, huma.StatusError) {
	header = strings.TrimSpace(header)
	if header == "" {
		return 0, huma.Error428PreconditionRequired("If-Match header is required")
	}
	s, ok := strings.CutPrefix(header, `"`)
	if ok {
		s, ok = strings.CutSuffix(s, `"`)
	}
	v, err := strconv.ParseUint(s, 10, 63)
	if !ok || err != nil || v == 0 {
		return 0, huma.Error400BadRequest(`If-Match must be a single strong ETag, e.g. "1"`)
	}
	return int64(v), nil
}

package pagination

import "errors"

// MinLimit is the smallest page size accepted for any collection.
const MinLimit = 1

var ErrInvalid = errors.New("pagination: invalid")

type Page struct {
	limit  int32
	offset int32
}

// NewPage validates limit against [MinLimit, maxLimit] and offset >= 0.
// maxLimit is chosen by each aggregate (e.g. actor.MaxPageLimit).
func NewPage(limit, offset, maxLimit int32) (Page, error) {
	if limit < MinLimit || limit > maxLimit || offset < 0 {
		return Page{}, ErrInvalid
	}
	return Page{limit: limit, offset: offset}, nil
}

func (p Page) Limit() int32      { return p.limit }
func (p Page) Offset() int32     { return p.offset }
func (p Page) HasPrevious() bool { return p.offset > 0 }
func (p Page) HasNext(total int64) bool {
	return int64(p.offset)+int64(p.limit) < total
}

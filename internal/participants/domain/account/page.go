package account

const (
	MinPageLimit = 1
	MaxPageLimit = 100
)

type Page struct {
	limit  int32
	offset int32
}

func NewPage(limit, offset int32) (Page, error) {
	if limit < MinPageLimit || limit > MaxPageLimit || offset < 0 {
		return Page{}, ErrInvalidPagination
	}
	return Page{limit: limit, offset: offset}, nil
}

func (p Page) Limit() int32      { return p.limit }
func (p Page) Offset() int32     { return p.offset }
func (p Page) HasPrevious() bool { return p.offset > 0 }
func (p Page) HasNext(total int64) bool {
	return int64(p.offset+p.limit) < total
}

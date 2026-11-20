// Package db provides all functionality for an in-memory store for links.
package db

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"sync"
	"time"
)

// Link represents a link inside the store.
type Link struct {
	URL       string
	Code      string
	Clicks    int
	CreatedAt time.Time
}

// LinksStore is an in-memory store for links.
type LinksStore struct {
	mu    sync.Mutex
	links map[string]Link
}

var (
	// ErrAlreadyExists is returned when a code already exists.
	ErrAlreadyExists = errors.New("already exists")
	// ErrNotFound is returned when a code is not found.
	ErrNotFound = errors.New("not found")
	// ErrInvalidPagination is returned an invalid limit or offset is provided.
	ErrInvalidPagination = errors.New("invalid pagination")
)

const codeLen = 7
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range n {
		b[i] = charset[rand.IntN(len(charset))]
	}
	return string(b)
}

// NewLinksStore setups an in-memory store for links.
func NewLinksStore() *LinksStore {
	return &LinksStore{links: map[string]Link{}}
}

// Create creates a new link to the db. If code is not provided it generates a random
// one with length [codeLen].
//
// If the code already exists it returns [ErrAlreadyExists].
func (s *LinksStore) Create(_ context.Context, url, code string) (Link, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if code == "" {
		code = randomString(codeLen)
	}
	if _, ok := s.links[code]; ok {
		return Link{}, ErrAlreadyExists
	}

	link := Link{URL: url, Code: code, CreatedAt: time.Now()}
	s.links[code] = link
	return link, nil
}

// Get gets the link that matches `code`.
//
// If the link does not exist it returns [ErrNotFound].
func (s *LinksStore) Get(_ context.Context, code string) (Link, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	link, ok := s.links[code]
	if !ok {
		return Link{}, ErrNotFound
	}
	return link, nil
}

// List lists the first limit links after offset. It sorts them by createdAt and uses
// code as a tie breaker.
func (s *LinksStore) List(_ context.Context, limit, offset int) ([]Link, error) {
	if limit < 0 {
		return nil, fmt.Errorf("%w: limit %d", ErrInvalidPagination, limit)
	}
	if offset < 0 {
		return nil, fmt.Errorf("%w: offset %d", ErrInvalidPagination, offset)
	}

	s.mu.Lock()
	links := make([]Link, 0, len(s.links))
	for _, l := range s.links {
		links = append(links, l)
	}
	s.mu.Unlock()

	slices.SortFunc(links, func(a, b Link) int {
		return cmp.Or(
			a.CreatedAt.Compare(b.CreatedAt),
			cmp.Compare(a.Code, b.Code),
		)
	})

	res := links[min(offset, len(links)):min(offset+limit, len(links))]
	return slices.Clone(res), nil
}

// Delete deletes the link from db.
//
// If the link does not exist it returns [ErrNotFound].
func (s *LinksStore) Delete(_ context.Context, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.links[code]; !ok {
		return ErrNotFound
	}

	delete(s.links, code)
	return nil
}

// Visit increase the number of Clicks of a link.
//
// If the link does not exist it returns [ErrNotFound].
func (s *LinksStore) Visit(_ context.Context, code string) (Link, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	link, ok := s.links[code]
	if !ok {
		return Link{}, ErrNotFound
	}

	link.Clicks++
	s.links[code] = link
	return link, nil
}

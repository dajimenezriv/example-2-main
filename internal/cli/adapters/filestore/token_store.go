package filestore

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"exercise/internal/cli/domain/auth"

	"golang.org/x/oauth2"
)

const (
	appDirName    = "data"
	tokenFileName = "token.json"
	dirPerm       = 0o700
	filePerm      = 0o600
)

type TokenStore struct {
	path string
}

func NewTokenStore(path string) *TokenStore {
	return &TokenStore{path: path}
}

func DefaultTokenStore() (*TokenStore, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}
	return NewTokenStore(filepath.Join(dir, appDirName, tokenFileName)), nil
}

func (s *TokenStore) Path() string { return s.path }

func (s *TokenStore) Load() (*oauth2.Token, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, auth.ErrNotLoggedIn
	}
	if err != nil {
		return nil, fmt.Errorf("read token file: %w", err)
	}
	var tok oauth2.Token
	if err := json.Unmarshal(data, &tok); err != nil {
		return nil, fmt.Errorf("decode token file %s: %w", s.path, err)
	}
	return &tok, nil
}

func (s *TokenStore) Save(tok *oauth2.Token) error {
	//nolint:gosec // Ignore marshaling an access token.
	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return fmt.Errorf("encode token: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), dirPerm); err != nil {
		return fmt.Errorf("create token dir: %w", err)
	}
	// WriteFile applies filePerm only when creating: tighten an existing file before the token goes in.
	if err := os.Chmod(s.path, filePerm); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("restrict token file: %w", err)
	}
	if err := os.WriteFile(s.path, data, filePerm); err != nil {
		return fmt.Errorf("write token file: %w", err)
	}
	return nil
}

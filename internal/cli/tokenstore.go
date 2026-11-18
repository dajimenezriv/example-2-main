package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

func DefaultTokenStore() (*TokenStore, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}
	return &TokenStore{path: filepath.Join(dir, appDirName, tokenFileName)}, nil
}

func (s *TokenStore) Path() string { return s.path }

func (s *TokenStore) Save(tok *oauth2.Token) error {
	//nolint:gosec // Ignore marshaling an access token.
	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return fmt.Errorf("encode token: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.path), dirPerm); err != nil {
		return fmt.Errorf("create token dir: %w", err)
	}
	if err := os.WriteFile(s.path, data, filePerm); err != nil {
		return fmt.Errorf("write token file: %w", err)
	}
	return nil
}

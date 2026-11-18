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
	// Tokens are credentials: only the current user may read them.
	dirPerm  = 0o700
	filePerm = 0o600
)

// TokenStore keeps the OAuth2 tokens in a JSON file.
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

// Path is where the tokens are stored.
func (s *TokenStore) Path() string { return s.path }

// Save writes the tokens to the file, creating it 0600 if needed.
func (s *TokenStore) Save(tok *oauth2.Token) error {
	//nolint:gosec // G117: persisting the tokens is this store's purpose; the file is 0600 in a 0700 dir.
	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return fmt.Errorf("encode token: %w", err)
	}
	if mkErr := os.MkdirAll(filepath.Dir(s.path), dirPerm); mkErr != nil {
		return fmt.Errorf("create token dir: %w", mkErr)
	}
	if wErr := os.WriteFile(s.path, data, filePerm); wErr != nil {
		return fmt.Errorf("write token file: %w", wErr)
	}
	return nil
}

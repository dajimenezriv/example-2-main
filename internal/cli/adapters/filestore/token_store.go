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
	//nolint:gosec // G117: persisting the token is this store's job; the file is 0600.
	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return fmt.Errorf("encode token: %w", err)
	}
	if err = os.MkdirAll(filepath.Dir(s.path), dirPerm); err != nil {
		return fmt.Errorf("create token dir: %w", err)
	}
	// Temp file + rename: refresh tokens rotate, so a torn write would lose the only valid one.
	// Same dir keeps the rename atomic; CreateTemp makes it 0600 (replacing any looser old file).
	tmp, err := os.CreateTemp(filepath.Dir(s.path), "."+filepath.Base(s.path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp token file: %w", err)
	}
	if err = writeAndClose(tmp, data); err != nil {
		return errors.Join(fmt.Errorf("write token file: %w", err), os.Remove(tmp.Name()))
	}
	if err = os.Rename(tmp.Name(), s.path); err != nil {
		return errors.Join(fmt.Errorf("replace token file: %w", err), os.Remove(tmp.Name()))
	}
	return nil
}

func writeAndClose(f *os.File, data []byte) error {
	_, err := f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	return errors.Join(err, f.Close())
}

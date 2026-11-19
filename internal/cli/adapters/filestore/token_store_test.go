package filestore

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"exercise/internal/cli/domain/auth"

	"golang.org/x/oauth2"
)

func newToken(access string) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  access,
		TokenType:    "Bearer",
		RefreshToken: "refresh-" + access,
		// No monotonic clock reading and UTC, so it survives the JSON round trip unchanged.
		Expiry: time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC),
	}
}

func perm(t *testing.T, path string) fs.FileMode {
	t.Helper()
	info, err := os.Stat(path)
	require.NoError(t, err)
	return info.Mode().Perm()
}

func TestTokenStore_SaveAndLoad(t *testing.T) {
	tests := []struct {
		name  string
		saves []*oauth2.Token
		want  *oauth2.Token
	}{
		{
			name:  "round trip",
			saves: []*oauth2.Token{newToken("a")},
			want:  newToken("a"),
		},
		{
			// A refresh rotates the token: the latest one must win.
			name:  "later save overwrites",
			saves: []*oauth2.Token{newToken("a"), newToken("b")},
			want:  newToken("b"),
		},
		{
			name:  "token without refresh token or expiry",
			saves: []*oauth2.Token{{AccessToken: "a", TokenType: "Bearer"}},
			want:  &oauth2.Token{AccessToken: "a", TokenType: "Bearer"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewTokenStore(filepath.Join(t.TempDir(), "token.json"))

			for _, tok := range tt.saves {
				require.NoError(t, store.Save(tok))
			}
			got, err := store.Load()

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTokenStore_Permissions(t *testing.T) {
	t.Run("creates missing dirs as 0700 and the file as 0600", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "data", "nested")
		path := filepath.Join(dir, "token.json")

		require.NoError(t, NewTokenStore(path).Save(newToken("a")))

		assert.Equal(t, fs.FileMode(0o700), perm(t, dir))
		assert.Equal(t, fs.FileMode(0o600), perm(t, path))
	})

	t.Run("tightens an existing world-readable file to 0600", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "token.json")
		require.NoError(t, os.WriteFile(path, []byte("{}"), 0o644)) //nolint:gosec // G306: the loose mode is the point of the test.

		require.NoError(t, NewTokenStore(path).Save(newToken("a")))

		assert.Equal(t, fs.FileMode(0o600), perm(t, path))
	})
}

func TestTokenStore_LoadErrors(t *testing.T) {
	tests := []struct {
		name        string
		content     *string
		wantErrIs   error
		wantMessage string
	}{
		{
			name:      "no file means not logged in",
			wantErrIs: auth.ErrNotLoggedIn,
		},
		{
			name:        "corrupt file is an error, not 'not logged in'",
			content:     new("{not json"),
			wantMessage: "decode token file",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "token.json")
			if tt.content != nil {
				require.NoError(t, os.WriteFile(path, []byte(*tt.content), 0o600))
			}

			got, err := NewTokenStore(path).Load()

			require.Error(t, err)
			assert.Nil(t, got)
			if tt.wantErrIs != nil {
				require.ErrorIs(t, err, tt.wantErrIs)
			} else {
				require.NotErrorIs(t, err, auth.ErrNotLoggedIn)
				assert.Contains(t, err.Error(), tt.wantMessage)
			}
		})
	}
}

func TestTokenStore_SaveFailsWhenParentIsAFile(t *testing.T) {
	parent := filepath.Join(t.TempDir(), "not-a-dir")
	require.NoError(t, os.WriteFile(parent, nil, 0o600))

	err := NewTokenStore(filepath.Join(parent, "token.json")).Save(newToken("a"))

	require.ErrorContains(t, err, "create token dir")
}

func TestDefaultTokenStore_UsesDataDirUnderWorkingDir(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	store, err := DefaultTokenStore()

	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "data", "token.json"), store.Path())
}

package listactors

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"exercise/internal/cli/adapters/commands/cmdtest"
	"exercise/internal/cli/domain/actor"
	"exercise/internal/cli/domain/actor/actortest"
)

var (
	actorID1   = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	actorID2   = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	actorID3   = uuid.MustParse("00000000-0000-0000-0000-000000000003")
	accountID1 = uuid.MustParse("00000000-0000-0000-0000-000000000004")
	accountID2 = uuid.MustParse("00000000-0000-0000-0000-000000000005")
	accountID3 = uuid.MustParse("00000000-0000-0000-0000-000000000006")
)

func newSUT(t *testing.T) actor.Repository {
	t.Helper()
	return actortest.NewFakeRepository(
		actor.Actor{ID: actorID1, TaxID: "000000000", Type: actor.TypeIssuer, Version: 1, Accounts: []actor.Account{
			{ID: accountID1, DltAccountID: "dlt1", PublicKey: "pub1"},
			{ID: accountID2, DltAccountID: "dlt2", PublicKey: "pub2"},
		}},
		actor.Actor{ID: actorID2, TaxID: "000000001", Type: actor.TypeInvestor, Version: 3, Accounts: []actor.Account{
			{ID: accountID3, DltAccountID: "dlt3", PublicKey: "pub3"},
		}},
		actor.Actor{ID: actorID3, TaxID: "000000002", Type: actor.TypeCustom, Version: 1},
	)
}

func list(ctx context.Context, t *testing.T, repo actor.Repository, args ...string) cmdtest.Result {
	t.Helper()
	return cmdtest.Run(ctx, t, NewCommand(repo).Register, append([]string{"list"}, args...)...)
}

const (
	header = "ID                                    TAX ID     TYPE      VERSION  ACCOUNTS\n"
	row1   = "00000000-0000-0000-0000-000000000001  000000000  issuer    1        2\n"
	row2   = "00000000-0000-0000-0000-000000000002  000000001  investor  3        1\n"
	row3   = "00000000-0000-0000-0000-000000000003  000000002  custom    1        0\n" // Only with row2 present (sets TYPE width).
)

var (
	actor1JSON = fmt.Sprintf(`{"id": %q, "taxId": "000000000", "type": "issuer", "version": 1, "accounts": [
		{"id": %q, "dltAccountId": "dlt1", "publicKey": "pub1"},
		{"id": %q, "dltAccountId": "dlt2", "publicKey": "pub2"}
	]}`, actorID1, accountID1, accountID2)
	actor2JSON = fmt.Sprintf(`{"id": %q, "taxId": "000000001", "type": "investor", "version": 3, "accounts": [
		{"id": %q, "dltAccountId": "dlt3", "publicKey": "pub3"}
	]}`, actorID2, accountID3)
	actor3JSON = fmt.Sprintf(`{"id": %q, "taxId": "000000002", "type": "custom", "version": 1, "accounts": []}`, actorID3)
)

func page(hasNext, hasPrevious bool, actors ...string) string {
	return fmt.Sprintf(
		`{"actors": [%s], "totalCount": 3, "hasNext": %t, "hasPrevious": %t}`,
		strings.Join(actors, ", "), hasNext, hasPrevious,
	)
}

func TestListActors_Table(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantStdout string
		wantStderr string
	}{
		{
			name:       "defaults",
			wantStdout: header + row1 + row2 + row3,
			wantStderr: "\nShowing 1-3 of 3.\n",
		},
		{
			name:       "explicit table output",
			args:       []string{"-o", "table"},
			wantStdout: header + row1 + row2 + row3,
			wantStderr: "\nShowing 1-3 of 3.\n",
		},
		{
			name:       "first page suggests the next offset",
			args:       []string{"--limit", "2"},
			wantStdout: header + row1 + row2,
			wantStderr: "\nShowing 1-2 of 3. Next page: --offset 2\n",
		},
		{
			name:       "middle page",
			args:       []string{"--limit", "1", "--offset", "1"},
			wantStdout: header + row2,
			wantStderr: "\nShowing 2-2 of 3. Next page: --offset 2\n",
		},
		{
			name: "last page has no next offset",
			args: []string{"--limit", "2", "--offset", "2"},
			// Columns are sized per page: without "investor", TYPE is narrower than in header.
			wantStdout: "ID                                    TAX ID     TYPE    VERSION  ACCOUNTS\n" +
				"00000000-0000-0000-0000-000000000003  000000002  custom  1        0\n",
			wantStderr: "\nShowing 3-3 of 3.\n",
		},
		{
			name:       "offset past end prints only the header and no footer",
			args:       []string{"--offset", "10"},
			wantStdout: "ID  TAX ID  TYPE  VERSION  ACCOUNTS\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := list(t.Context(), t, newSUT(t), tt.args...)

			require.NoError(t, res.Err)
			assert.Equal(t, tt.wantStdout, res.Stdout)
			assert.Equal(t, tt.wantStderr, res.Stderr)
		})
	}
}

func TestListActors_JSON(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "all actors",
			args: []string{"-o", "json"},
			want: page(false, false, actor1JSON, actor2JSON, actor3JSON),
		},
		{
			name: "long flag",
			args: []string{"--output", "json", "--limit", "1", "--offset", "1"},
			want: page(true, true, actor2JSON),
		},
		{
			name: "offset past end",
			args: []string{"-o", "json", "--offset", "10"},
			want: page(false, true),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := list(t.Context(), t, newSUT(t), tt.args...)

			require.NoError(t, res.Err)
			assert.JSONEq(t, tt.want, res.Stdout)
			// No footer: stdout must stay parseable and stderr quiet for scripts.
			assert.Empty(t, res.Stderr)
		})
	}
}

func TestListActors_Errors(t *testing.T) {
	canceled, cancel := context.WithCancel(t.Context())
	cancel()

	tests := []struct {
		name    string
		ctx     context.Context
		args    []string
		wantErr string
	}{
		{
			name:    "repository error is wrapped with the command",
			ctx:     canceled,
			wantErr: "list actors: context canceled",
		},
		{
			name:    "unknown output format",
			ctx:     t.Context(),
			args:    []string{"-o", "yaml"},
			wantErr: `list actors: invalid output "yaml": want table or json`,
		},
		{
			name:    "non-numeric limit",
			ctx:     t.Context(),
			args:    []string{"--limit", "abc"},
			wantErr: `invalid argument "abc" for "--limit" flag`,
		},
		{
			name:    "positional argument",
			ctx:     t.Context(),
			args:    []string{"extra"},
			wantErr: `unknown command "extra" for "exercise list"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := list(tt.ctx, t, newSUT(t), tt.args...)

			require.Error(t, res.Err)
			assert.Contains(t, res.Err.Error(), tt.wantErr)
			assert.Empty(t, res.Stdout)
		})
	}
}

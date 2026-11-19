package createactor

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"exercise/internal/cli/adapters/commands/cmdtest"
	"exercise/internal/cli/domain/actor"
	"exercise/internal/cli/domain/actor/actortest"
)

func create(ctx context.Context, t *testing.T, repo actor.Repository, args ...string) cmdtest.Result {
	t.Helper()
	return cmdtest.Run(ctx, t, NewCommand(repo).Register, append([]string{"create"}, args...)...)
}

func storedActors(t *testing.T, repo actor.Repository) []actor.Actor {
	t.Helper()
	page, err := repo.FindAll(t.Context(), 10, 0)
	require.NoError(t, err)
	return page.Actors
}

// The fake assigns a random ID, so it is read back from the repo; this also proves the actor was stored.
func onlyStoredActor(t *testing.T, repo actor.Repository) actor.Actor {
	t.Helper()
	actors := storedActors(t, repo)
	require.Len(t, actors, 1)
	return actors[0]
}

func TestCreateActor_Table(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantActor  actor.Actor
		wantStdout string // %s is the created actor's ID.
	}{
		{
			name:      "defaults to table",
			args:      []string{"--tax-id", "123456789", "--type", "issuer"},
			wantActor: actor.Actor{TaxID: "123456789", Type: actor.TypeIssuer, Version: 1},
			wantStdout: "ID                                    TAX ID     TYPE    VERSION\n" +
				"%s  123456789  issuer  1\n",
		},
		{
			name:      "explicit table output, wider type column",
			args:      []string{"--tax-id", "123456789", "--type", "settlement_operator", "-o", "table"},
			wantActor: actor.Actor{TaxID: "123456789", Type: actor.TypeSettlementOperator, Version: 1},
			wantStdout: "ID                                    TAX ID     TYPE                 VERSION\n" +
				"%s  123456789  settlement_operator  1\n",
		},
		{
			// The API is the source of truth for validation (its 422 lists every invalid field).
			name:      "values are passed through unvalidated",
			args:      []string{"--tax-id", "12", "--type", "nope"},
			wantActor: actor.Actor{TaxID: "12", Type: "nope", Version: 1},
			wantStdout: "ID                                    TAX ID  TYPE  VERSION\n" +
				"%s  12      nope  1\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := actortest.NewFakeRepository()

			res := create(t.Context(), t, repo, tt.args...)

			require.NoError(t, res.Err)
			got := onlyStoredActor(t, repo)
			tt.wantActor.ID = got.ID
			assert.Equal(t, tt.wantActor, got)
			assert.Equal(t, fmt.Sprintf(tt.wantStdout, got.ID), res.Stdout)
			assert.Empty(t, res.Stderr)
		})
	}
}

func TestCreateActor_JSON(t *testing.T) {
	repo := actortest.NewFakeRepository()

	res := create(t.Context(), t, repo, "--tax-id", "123456789", "--type", "investor", "--output", "json")

	require.NoError(t, res.Err)
	got := onlyStoredActor(t, repo)
	assert.JSONEq(t,
		fmt.Sprintf(`{"id": %q, "taxId": "123456789", "type": "investor", "version": 1}`, got.ID),
		res.Stdout)
	assert.Empty(t, res.Stderr)
}

func TestCreateActor_Errors(t *testing.T) {
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
			args:    []string{"--tax-id", "123456789", "--type", "issuer"},
			wantErr: "create actor: context canceled",
		},
		{
			// Checked before calling the repo: a typo must not create the actor and then fail.
			name:    "unknown output format",
			ctx:     t.Context(),
			args:    []string{"--tax-id", "123456789", "--type", "issuer", "-o", "yaml"},
			wantErr: `create actor: invalid output "yaml": want table or json`,
		},
		{
			name:    "missing tax ID",
			ctx:     t.Context(),
			args:    []string{"--type", "issuer"},
			wantErr: `required flag(s) "tax-id" not set`,
		},
		{
			name:    "missing type",
			ctx:     t.Context(),
			args:    []string{"--tax-id", "123456789"},
			wantErr: `required flag(s) "type" not set`,
		},
		{
			name:    "no flags",
			ctx:     t.Context(),
			wantErr: `required flag(s) "tax-id", "type" not set`,
		},
		{
			name:    "positional argument",
			ctx:     t.Context(),
			args:    []string{"123456789", "--tax-id", "123456789", "--type", "issuer"},
			wantErr: `unknown command "123456789" for "exercise create"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := actortest.NewFakeRepository()

			res := create(tt.ctx, t, repo, tt.args...)

			require.Error(t, res.Err)
			assert.Contains(t, res.Err.Error(), tt.wantErr)
			assert.Empty(t, res.Stdout)
			assert.Empty(t, storedActors(t, repo))
		})
	}
}

func TestCreateActor_TypeCompletion(t *testing.T) {
	// __complete is cobra's hidden command behind shell completion; ":4" is ShellCompDirectiveNoFileComp.
	res := cmdtest.Run(t.Context(), t, NewCommand(actortest.NewFakeRepository()).Register, "__complete", "create", "--type", "")

	require.NoError(t, res.Err)
	assert.Equal(t, "issuer\ninvestor\nplatform_admin\nmarket_operator\nsettlement_operator\ncustom\n:4\n", res.Stdout)
}

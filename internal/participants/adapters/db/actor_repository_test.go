package db

import (
	"context"
	"errors"
	"testing"
	"uuid"

	"exercise/internal/participants/adapters/db/sqlcgen"
	"exercise/internal/participants/domain/account"
	"exercise/internal/participants/domain/account/accounttest"
	"exercise/internal/participants/domain/actor"
	"exercise/internal/participants/domain/actor/actortest"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActorRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test")
	}

	q := sqlcgen.New(testPool)

	actortest.RunRepositoryContract(t, func(t *testing.T) actor.Repository {
		t.Helper()
		if _, err := testPool.Exec(t.Context(), "TRUNCATE actors CASCADE"); err != nil {
			t.Fatalf("truncate actors: %v", err)
		}
		return NewActorRepository(q)
	})
}

type auditRow struct {
	EntityType string
	EntityID   uuid.UUID
	DeletedBy  *string
	Data       map[string]any
}

func auditRows(t *testing.T) []auditRow {
	t.Helper()
	rows, err := testPool.Query(t.Context(),
		"SELECT entity_type, entity_id, deleted_by, data FROM deletion_audit ORDER BY id")
	require.NoError(t, err)
	defer rows.Close()

	var res []auditRow
	for rows.Next() {
		var r auditRow
		require.NoError(t, rows.Scan(&r.EntityType, &r.EntityID, &r.DeletedBy, &r.Data))
		res = append(res, r)
	}
	require.NoError(t, rows.Err())
	return res
}

type cascadeFixture struct {
	target, other  uuid.UUID
	targetAccounts []*account.Account
}

func TestActorRepositoryDeleteCascade(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test")
	}

	q := sqlcgen.New(testPool)
	actors := NewActorRepository(q)
	accounts := NewAccountRepository(q)

	setup := func(t *testing.T) cascadeFixture {
		t.Helper()
		_, err := testPool.Exec(t.Context(), "TRUNCATE actors, deletion_audit CASCADE")
		require.NoError(t, err, "truncate")

		f := cascadeFixture{target: seedActor(t, q), other: seedActor(t, q)}
		for _, actorID := range []uuid.UUID{f.target, f.target, f.other} {
			a := accounttest.NewAccount(t, actorID)
			require.NoError(t, accounts.Save(t.Context(), a))
			if actorID == f.target {
				f.targetAccounts = append(f.targetAccounts, a)
			}
		}
		return f
	}

	accountsOf := func(t *testing.T, actorID uuid.UUID) []sqlcgen.Account {
		t.Helper()
		got, err := q.FindAccountsByActorIDs(t.Context(), []uuid.UUID{actorID})
		require.NoError(t, err)
		return got
	}

	t.Run("Delete cascades to the actor's accounts only", func(t *testing.T) {
		f := setup(t)

		require.NoError(t, actors.Delete(t.Context(), f.target, 1))

		assert.Empty(t, accountsOf(t, f.target))
		assert.Len(t, accountsOf(t, f.other), 1)
	})

	t.Run("Delete audits the actor and every cascaded account", func(t *testing.T) {
		f := setup(t)
		target := f.target

		require.NoError(t, actors.Delete(t.Context(), target, 1))

		got := auditRows(t)
		require.Len(t, got, 3)

		byID := map[uuid.UUID]auditRow{}
		for _, r := range got {
			byID[r.EntityID] = r
		}

		act := byID[target]
		assert.Equal(t, "actor", act.EntityType)
		assert.Equal(t, target.String(), act.Data["id"])
		assert.Nil(t, act.DeletedBy)

		for _, a := range f.targetAccounts {
			r, ok := byID[a.ID()]
			require.True(t, ok, "missing audit row for account %s", a.ID())
			assert.Equal(t, "account", r.EntityType)
			assert.Equal(t, target.String(), r.Data["actor_id"])
			assert.Equal(t, a.PrivateKey(), r.Data["private_key"])
			assert.Equal(t, a.PublicKey(), r.Data["public_key"])
		}
	})

	t.Run("Delete records app.principal as deleted_by", func(t *testing.T) {
		f := setup(t)

		tx, err := testPool.Begin(t.Context())
		require.NoError(t, err)
		t.Cleanup(func() {
			// t.Context is already canceled during cleanup.
			if rbErr := tx.Rollback(context.Background()); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
				t.Errorf("rollback: %v", rbErr)
			}
		})
		_, err = tx.Exec(t.Context(), "SELECT set_config('app.principal', 'alice', true)")
		require.NoError(t, err)
		require.NoError(t, NewActorRepository(sqlcgen.New(tx)).Delete(t.Context(), f.target, 1))
		require.NoError(t, tx.Commit(t.Context()))

		got := auditRows(t)
		require.Len(t, got, 3)
		for _, r := range got {
			require.NotNil(t, r.DeletedBy)
			assert.Equal(t, "alice", *r.DeletedBy)
		}
	})

	t.Run("Failed delete writes no audit rows and keeps accounts", func(t *testing.T) {
		f := setup(t)

		err := actors.Delete(t.Context(), f.target, 2)
		require.ErrorIs(t, err, actor.ErrVersionConflict)

		assert.Empty(t, auditRows(t))
		assert.Len(t, accountsOf(t, f.target), 2)
	})
}

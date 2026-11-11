package db

import (
	"context"
	"errors"
	"fmt"

	"exercise/internal/participants/adapters/db/sqlcgen"
	"exercise/internal/participants/app/uow"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UnitOfWork struct {
	pool *pgxpool.Pool
}

func NewUnitOfWork(pool *pgxpool.Pool) *UnitOfWork {
	return &UnitOfWork{pool: pool}
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(ctx context.Context, repos uow.Repositories) error) (err error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		// WithoutCancel lets reach the database even when ctx was the reason fn failed.
		rbErr := tx.Rollback(context.WithoutCancel(ctx))
		if rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			err = errors.Join(err, fmt.Errorf("rollback tx: %w", rbErr))
		}
	}()

	q := sqlcgen.New(tx)
	repos := uow.Repositories{
		Accounts: NewAccountRepository(q),
		Actors:   NewActorRepository(q),
	}

	if err := fn(ctx, repos); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

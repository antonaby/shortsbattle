package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/antonaby/shortsbattle/game-server/internal/db/q"
)

type TxFunc func(context.Context, pgx.Tx) error
type TxFuncWithValue[T any] func(context.Context, pgx.Tx) (T, error)

type TxManager interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Querier(pgx.Tx) q.Querier
}

type DbManager struct {
	Pool *pgxpool.Pool
}

func NewDbManager(ctx context.Context) (*DbManager, error) {
	dsn := os.Getenv("DATABASE_URL")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	manager := &DbManager{
		Pool: pool,
	}

	return manager, nil
}

func (m *DbManager) Close() {
	m.Pool.Close()
}

func (m *DbManager) Querier(tx pgx.Tx) q.Querier {
	queries := q.New(tx)
	return queries
}

func (m *DbManager) Begin(ctx context.Context) (pgx.Tx, error) {
	return m.Pool.Begin(ctx)
}

func WithTx(ctx context.Context, txm TxManager, fn TxFunc) error {
	tx, err := txm.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(ctx, tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func WithTxValue[T any](ctx context.Context, txm TxManager, fn TxFuncWithValue[T]) (T, error) {
	var zero T

	tx, err := txm.Begin(ctx)
	if err != nil {
		return zero, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	result, err := fn(ctx, tx)
	if err != nil {
		_ = tx.Rollback(ctx)
		return zero, err
	}

	return result, tx.Commit(ctx)
}

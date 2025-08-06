package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxFunc func(pgx.Tx) error
type TxValueFunc[T any] func(pgx.Tx) (T, error)

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

func (m *DbManager) Querier() Querier {
	queries := New(m.Pool)
	return queries
}

func WithTransaction(ctx context.Context, m *DbManager, fn TxFunc) error {
	tx, err := m.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func WithValueTransaction[T any](ctx context.Context, m *DbManager, fn TxValueFunc[T]) (T, error) {
	var zero T

	tx, err := m.Pool.Begin(ctx)
	if err != nil {
		return zero, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	result, err := fn(tx)
	if err != nil {
		_ = tx.Rollback(ctx)
		return zero, err
	}

	return result, tx.Commit(ctx)
}

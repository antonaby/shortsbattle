package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
)

type TxFunc func(context.Context, pgx.Tx) error
type TxFuncWithValue[T any] func(context.Context, pgx.Tx) (T, error)

type TxManager interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Querier(pgx.Tx) qg.Querier
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

func (m *DbManager) Querier(tx pgx.Tx) qg.Querier {
	queries := qg.New(tx)
	return queries
}

func (m *DbManager) Begin(ctx context.Context) (pgx.Tx, error) {
	return m.Pool.Begin(ctx)
}

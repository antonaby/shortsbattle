package db

import (
	"context"
	"errors"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
)

var ErrorDbUrlNotDefined = errors.New("DATABASE_URL not defined")

type TxManager interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Querier(pgx.Tx) qg.Querier
}

type DbManager struct {
	Pool *pgxpool.Pool
}

func NewDbManager(ctx context.Context, dsn *string) (*DbManager, error) {
	var dbUrl string
	if dsn != nil {
		dbUrl = *dsn
	} else {
		dbUrl = os.Getenv("DATABASE_URL")
		if len(dbUrl) == 0 {
			return nil, ErrorDbUrlNotDefined
		}
	}

	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		return nil, err
	}

	manager := &DbManager{
		Pool: pool,
	}

	return manager, nil
}

func (m *DbManager) Ping(ctx context.Context) error {
	return m.Pool.Ping(ctx)
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

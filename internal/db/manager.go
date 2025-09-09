package db

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pressly/goose/v3"

	"github.com/antonaby/shortsbattle/game-server/db/migrations"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrorDbUrlNotDefined = errors.New("DATABASE_URL not defined")

type TxManager interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Querier(pgx.Tx) qg.Querier
}

type DbManager struct {
	Pool *pgxpool.Pool
	DSN  string
}

func NewDbManager(dsn *string) (*DbManager, error) {
	var dbUrl string
	if dsn != nil {
		dbUrl = *dsn
	} else {
		dbUrl = os.Getenv("DATABASE_URL")
		if len(dbUrl) == 0 {
			return nil, ErrorDbUrlNotDefined
		}
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	pool, err := pgxpool.New(ctx, dbUrl)
	if err != nil {
		return nil, err
	}

	manager := &DbManager{
		Pool: pool,
		DSN:  dbUrl,
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

func (m *DbManager) Migrate() error {
	db, err := sql.Open("pgx", m.DSN)
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetBaseFS(migrations.Files)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db, ".")
}

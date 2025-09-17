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
	"github.com/rs/zerolog/log"
)

type QueryTracer struct {
}

func (l *QueryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	log.Debug().
		Str("query", data.SQL).
		Any("args", data.Args).
		Msg("query_start")

	return ctx
}

func (l *QueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
	if data.Err != nil {
		log.Error().
			Err(data.Err).
			Str("command_tag", data.CommandTag.String()).
			Msg("query_end")
		return
	}

	log.Debug().
		Str("command_tag", data.CommandTag.String()).
		Msg("query_end")
}

var ErrorDbUrlNotDefined = errors.New("DATABASE_URL not defined")

func GetDbDSN() (string, error) {
	dbUrl := os.Getenv("DATABASE_URL")
	if len(dbUrl) == 0 {
		return "", ErrorDbUrlNotDefined
	}

	return dbUrl, nil
}

type TxManager interface {
	Begin(ctx context.Context) (pgx.Tx, error)
	Querier(pgx.Tx) qg.Querier
}

type DbManager struct {
	Pool *pgxpool.Pool
	DSN  string
}

func NewDbManager(dsn string) (*DbManager, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	// TODO: add connection params, like max/min connections
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// TODO: add flag for enabling/disabling logger
	config.ConnConfig.Tracer = &QueryTracer{}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	manager := &DbManager{
		Pool: pool,
		DSN:  dsn,
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

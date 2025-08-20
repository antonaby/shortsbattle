package db

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

/*
23000 — integrity_constraint_violation (generic)
23001 — restrict_violation
23502 — not_null_violation
23503 — foreign_key_violation
23505 — unique_violation
23514 — check_violation
23P01 — exclusion_violation
*/
func IsClass23(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && strings.HasPrefix(pgErr.Code, "23")
}

func IsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
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

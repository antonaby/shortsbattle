package db

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

/*
PostgreSQL Error Class 23 — Integrity Constraint Violation
Reference: https://www.postgresql.org/docs/current/errcodes-appendix.html

23000 — integrity_constraint_violation
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

/*
PostgreSQL Error Class 22 — Data Exception
Reference: https://www.postgresql.org/docs/current/errcodes-appendix.html

22001 — string_data_right_truncation
22003 — numeric_value_out_of_range
22007 — invalid_datetime_format
22008 — datetime_field_overflow
22012 — division_by_zero
22018 — invalid_character_value_for_cast
22021 — character_not_in_repertoire
22023 — invalid_parameter_value
22025 — invalid_escape_sequence
2202E — array_subscript_error
22P01 — floating_point_exception
22P02 — invalid_text_representation
22P03 — invalid_binary_representation
22P04 — bad_copy_file_format
22P05 — untranslatable_character
*/
func IsClass22(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && strings.HasPrefix(pgErr.Code, "22")
}

func IsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func ToPgInterval(time time.Duration) pgtype.Interval {
	return pgtype.Interval{
		Microseconds: int64(time.Microseconds()),
		Days:         0,
		Months:       0,
		Valid:        true,
	}
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

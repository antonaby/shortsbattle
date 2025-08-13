package utils

import (
	"errors"
	"strings"

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
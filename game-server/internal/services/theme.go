package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/jackc/pgx/v5"
)

type ThemeServiceErrorCode int

const (
	TSErrDbError = iota
	TSErrNotFound
)

type ThemeServiceError struct {
	Code    ThemeServiceErrorCode
	Message string
	Cause   error
}

func (e ThemeServiceError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("Error %d: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

func (e ThemeServiceError) Unwrap() error {
	return e.Cause
}

type ThemeService struct {
	txm db.TxManager
}

func NewThemeService(txm db.TxManager) *ThemeService {
	return &ThemeService{
		txm: txm,
	}
}

func (ts *ThemeService) CreateTheme(ctx context.Context, params db.CreateThemeParams) (*db.Theme, error) {
	return db.WithTxValue(ctx, ts.txm, func(ctx context.Context, tx pgx.Tx) (*db.Theme, error) {
		q := ts.txm.Querier(tx)
		theme, err := q.CreateTheme(ctx, params)
		if err != nil {
			return nil, ThemeServiceError{
				Code:    TSErrDbError,
				Message: "can't create theme",
				Cause:   err,
			}
		}

		return &theme, nil
	})
}

func (ts *ThemeService) ListAllThemes(ctx context.Context) ([]db.Theme, error) {
	return db.WithTxValue(ctx, ts.txm, func(ctx context.Context, tx pgx.Tx) ([]db.Theme, error) {
		q := ts.txm.Querier(tx)
		themes, err := q.ListAllThemes(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ThemeServiceError{
					Code:    TSErrNotFound,
					Message: "no themes found",
					Cause:   err,
				}
			}

			return nil, ThemeServiceError{
				Code:    TSErrDbError,
				Message: "can't list all themes",
				Cause:   err,
			}
		}

		return themes, nil
	})
}

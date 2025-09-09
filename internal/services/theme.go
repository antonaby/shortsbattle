package services

import (
	"context"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type ThemeService struct {
	txm db.TxManager
}

func NewThemeService(txm db.TxManager) *ThemeService {
	return &ThemeService{
		txm: txm,
	}
}

func (ts *ThemeService) CreateTheme(ctx context.Context, title string, description pgtype.Text) (*qg.Theme, error) {
	return db.WithTxValue(ctx, ts.txm, func(ctx context.Context, tx pgx.Tx) (*qg.Theme, error) {
		q := ts.txm.Querier(tx)
		theme, err := q.CreateTheme(ctx, title, description)
		if err != nil {
			return nil, common.ServiceError{
				Code:    common.ErrorDbUnknown,
				Message: "failed to create a theme",
				Cause:   err,
			}
		}

		return &theme, nil
	})
}

func (ts *ThemeService) ListAllThemes(ctx context.Context) ([]qg.Theme, error) {
	return db.WithTxValue(ctx, ts.txm, func(ctx context.Context, tx pgx.Tx) ([]qg.Theme, error) {
		q := ts.txm.Querier(tx)
		themes, err := q.ListAllThemes(ctx)
		if err != nil {
			return nil, common.ServiceError{
				Code:    common.ErrorDbUnknown,
				Message: "failed to list all themes",
				Cause:   err,
			}
		}

		if len(themes) == 0 {
			themes = []qg.Theme{}
		}

		return themes, nil
	})
}

func (ts *ThemeService) GetTheme(ctx context.Context, themeId int64) (*qg.Theme, error) {
	return db.WithTxValue(ctx, ts.txm, func(ctx context.Context, tx pgx.Tx) (*qg.Theme, error) {
		q := ts.txm.Querier(tx)
		theme, err := q.GetTheme(ctx, themeId)
		if err != nil {
			return nil, common.ServiceError{
				Code:    common.GetDbErrorCode(err),
				Message: "failed to get theme",
				Cause:   err,
			}
		}

		return &theme, nil
	})
}

func (ts *ThemeService) CreateRound(ctx context.Context, params qg.CreateRoundParams) (*qg.Round, error) {
	return db.WithTxValue(ctx, ts.txm, func(ctx context.Context, tx pgx.Tx) (*qg.Round, error) {
		q := ts.txm.Querier(tx)
		request, err := q.CreateRound(ctx, params)
		if err != nil {
			return nil, common.ServiceError{
				Code:    common.GetDbErrorCode(err),
				Message: "failed to create a video request",
				Cause:   err,
			}
		}

		return &request, nil
	})
}

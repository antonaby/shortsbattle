package services

import (
	"context"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/jackc/pgx/v5"
)

type ThemeService struct {
	txm db.TxManager
}

func NewThemeService(txm db.TxManager) *ThemeService {
	return &ThemeService{
		txm: txm,
	}
}

func (ts *ThemeService) CreateTheme(ctx context.Context, params qg.CreateThemeParams) (*qg.Theme, error) {
	return db.WithTxValue(ctx, ts.txm, func(ctx context.Context, tx pgx.Tx) (*qg.Theme, error) {
		q := ts.txm.Querier(tx)
		theme, err := q.CreateTheme(ctx, params)
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

func (ts *ThemeService) GetTheme(ctx context.Context, themeId int64) (*qg.GetThemeRow, error) {
	return db.WithTxValue(ctx, ts.txm, func(ctx context.Context, tx pgx.Tx) (*qg.GetThemeRow, error) {
		q := ts.txm.Querier(tx)
		theme, err := q.GetTheme(ctx, qg.GetThemeParams{ID: themeId})
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

func (ts *ThemeService) CreateVideoRequest(ctx context.Context, params qg.CreateVideoRequestParams) (*qg.VideoRequest, error) {
	return db.WithTxValue(ctx, ts.txm, func(ctx context.Context, tx pgx.Tx) (*qg.VideoRequest, error) {
		q := ts.txm.Querier(tx)
		request, err := q.CreateVideoRequest(ctx, params)
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

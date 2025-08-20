package services

import (
	"context"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
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
				Code:    common.ErrorDb,
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
				Code:    common.ErrorDb,
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

func (ts *ThemeService) GetTheme(ctx context.Context, themeId int64) (*models.ThemeExt, error) {
	return db.WithTxValue(ctx, ts.txm, func(ctx context.Context, tx pgx.Tx) (*models.ThemeExt, error) {
		q := ts.txm.Querier(tx)
		theme, err := ts.getTheme(ctx, q, qg.GetThemeParams{ID: themeId})
		if err != nil {
			return nil, err
		}

		requests, err := ts.getVideoRequests(ctx, q, qg.GetVideoRequestsParams{ThemeID: themeId})
		if err != nil {
			return nil, err
		}

		return &models.ThemeExt{
			ID:          theme.ID,
			Name:        theme.Name,
			Description: theme.Description,
			CreatedAt:   theme.CreatedAt,
			Requests:    requests,
		}, nil
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

func (ts *ThemeService) getTheme(ctx context.Context, q qg.Querier, params qg.GetThemeParams) (*qg.Theme, error) {
	theme, err := q.GetTheme(ctx, params)
	if err != nil {
		if db.IsNoRows(err) {
			return nil, common.ServiceError{
				Code:    common.ErrorNotFound,
				Message: "failed to get theme",
				Cause:   err,
			}
		}

		return nil, common.ServiceError{
			Code:    common.ErrorDb,
			Message: "failed to get theme",
			Cause:   err,
		}
	}

	return &theme, nil
}

func (ts *ThemeService) getVideoRequests(ctx context.Context, q qg.Querier, params qg.GetVideoRequestsParams) ([]qg.VideoRequest, error) {
	requests, err := q.GetVideoRequests(ctx, params)
	if err != nil {
		return nil, common.ServiceError{
			Code:    common.ErrorDb,
			Message: "failed to list video requests",
			Cause:   err,
		}
	}

	if len(requests) == 0 {
		requests = []qg.VideoRequest{}
	}

	return requests, nil
}

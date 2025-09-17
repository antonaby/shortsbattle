package services

import (
	"context"
	"fmt"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
)

func tmError(code common.ErrorCode, msg string, err error) error {
	return common.ServiceError{
		Code:    code,
		Message: fmt.Sprintf("theme service: %s", msg),
		Cause:   err,
	}
}

type ThemeService struct {
	txm db.TxManager
}

func NewThemeService(txm db.TxManager) *ThemeService {
	return &ThemeService{
		txm: txm,
	}
}

func (ts *ThemeService) CreateTheme(ctx context.Context, title string, description pgtype.Text, mode qg.GameMode) (*models.ThemeWithRounds, error) {
	return db.WithTxVQ(ctx, ts.txm, func(ctx context.Context, q qg.Querier) (*models.ThemeWithRounds, error) {
		theme, err := q.CreateTheme(ctx, qg.CreateThemeParams{
			Title:       title,
			Description: description,
			Mode:        mode,
		})
		if err != nil {
			return nil, tmError(common.GetDbErrorCode(err), "failed to create theme", err)
		}

		return &models.ThemeWithRounds{
			ID:          theme.ID,
			Title:       theme.Title,
			Description: theme.Description.String,
			Mode:        theme.Mode,
		}, nil
	})
}

func (ts *ThemeService) ListAllThemes(ctx context.Context) ([]models.ThemeWithRounds, error) {
	return db.WithTxVQ(ctx, ts.txm, func(ctx context.Context, q qg.Querier) ([]models.ThemeWithRounds, error) {
		themes, err := q.ListAllThemes(ctx)
		if err != nil {
			return nil, tmError(common.GetDbErrorCode(err), "failed to list themes", err)
		}

		if len(themes) == 0 {
			return []models.ThemeWithRounds{}, nil
		}

		result := make([]models.ThemeWithRounds, 0, len(themes))
		for _, t := range themes {
			result = append(result, models.ThemeWithRounds{
				ID:          t.ID,
				Title:       t.Title,
				Description: t.Description.String,
				Mode:        t.Mode,
			})
		}

		return result, nil
	})
}

func (ts *ThemeService) GetTheme(ctx context.Context, themeId int64) (*models.ThemeWithRounds, error) {
	return db.WithTxVQ(ctx, ts.txm, func(ctx context.Context, q qg.Querier) (*models.ThemeWithRounds, error) {
		theme, err := q.GetTheme(ctx, themeId)
		if err != nil {
			return nil, tmError(common.GetDbErrorCode(err), "failed to get theme", err)
		}

		roundsRaw, err := q.GetRounds(ctx, theme.ID)
		if err != nil {
			return nil, tmError(common.GetDbErrorCode(err), "failed to get rounds", err)
		}

		rounds := []models.ThemeRound{}
		for _, i := range roundsRaw {
			rounds = append(rounds, models.ThemeRound{
				RoundN:      i.RoundN,
				Title:       i.Title,
				Description: i.Description.String,
			})
		}

		return &models.ThemeWithRounds{
			ID:          theme.ID,
			Title:       theme.Title,
			Description: theme.Description.String,
			Mode:        theme.Mode,
			Rounds:      rounds,
		}, nil
	})
}

func (ts *ThemeService) CreateRound(ctx context.Context, params qg.CreateRoundParams) (*models.ThemeRound, error) {
	return db.WithTxVQ(ctx, ts.txm, func(ctx context.Context, q qg.Querier) (*models.ThemeRound, error) {
		round, err := q.CreateRound(ctx, params)
		if err != nil {
			return nil, tmError(common.GetDbErrorCode(err), "failed to create round", err)
		}

		return &models.ThemeRound{
			RoundN:      round.RoundN,
			Title:       round.Title,
			Description: round.Description.String,
		}, nil
	})
}

package services

import (
	"context"
	"fmt"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/jackc/pgx/v5"
)

type GameManager struct {
	config GameConfig
	txm    db.TxManager
}

func NewGameManager(txm db.TxManager, config GameConfig) *GameManager {
	return &GameManager{
		txm:    txm,
		config: config,
	}
}

func (gm *GameManager) JoinGame(ctx context.Context, themeId int64, playerId int64) (int64, error) {
	return db.WithTxValue(ctx, gm.txm, func(ctx context.Context, tx pgx.Tx) (int64, error) {
		q := gm.txm.Querier(tx)
		gameId, err := q.JoinGameForTheme(ctx, qg.JoinGameForThemeParams{
			PThemeID:           themeId,
			PPlayerID:          playerId,
			PLobbyState:        qg.GameStateLobby,
			PLobbyStateClosed:  db.ToPgInterval(gm.config.LobbyClosedBefore),
			PMaxPlayers:        gm.config.MaxPlayers,
			PNextStateChangeIn: db.ToPgInterval(gm.config.LobbyState),
		})

		if err != nil {
			if db.IsClass23(err) {
				return 0, common.ServiceError{
					Code:    common.ErrorDbNotFound,
					Message: "theme or player not found",
					Cause:   err,
				}
			}

			return 0, common.ServiceError{
				Code:    common.ErrorDbUnknown,
				Message: "filed to join game",
				Cause:   err,
			}
		}

		return gameId, nil
	})
}

func (gm *GameManager) GetGameDetailsForPlayer(ctx context.Context, gameId int64, playerId int64) (*models.GameDetails, error) {
	return db.WithTxValue(ctx, gm.txm, func(ctx context.Context, tx pgx.Tx) (*models.GameDetails, error) {
		q := gm.txm.Querier(tx)
		game, err := q.GetGameForPlayer(ctx, qg.GetGameForPlayerParams{ID: gameId, PlayerID: playerId})
		if err != nil {
			return nil, common.ServiceError{
				Code:    common.GetDbErrorCode(err),
				Message: "failed to fetch game",
				Cause:   err,
			}
		}

		return &models.GameDetails{
			GameID:            game.ID,
			MsgType:           models.GameDetailsMsg,
			State:             game.State,
			StateChangedAt:    game.StateChangedAt,
			NextStateChangeAt: game.NextStateChangeAt,
			RamaningTimeMs:    game.RemainingMs,
		}, nil
	})
}

func (gm *GameManager) AdvanceGame(ctx context.Context, gameId int64) (*models.GameUpdate, error) {
	return db.WithTxValue(ctx, gm.txm, func(ctx context.Context, tx pgx.Tx) (*models.GameUpdate, error) {
		q := gm.txm.Querier(tx)
		game, err := q.GetGameAndLock(ctx, qg.GetGameAndLockParams{ID: gameId})
		if err != nil {
			return nil, common.ServiceError{
				Code:    common.GetDbErrorCode(err),
				Message: "failed to fetch game",
				Cause:   err,
			}
		}

		switch game.State {
		case qg.GameStateLobby:
			return gm.handleLobby(ctx, &game, q)
		case qg.GameStateSubmitting:
			return gm.handleSubmitting(ctx, &game, q)
		case qg.GameStateWatching:
			return gm.handleWathching(ctx, &game, q)
		default:
			return nil, common.ServiceError{
				Code:    common.ErrorWrongGameState,
				Message: fmt.Sprintf("wrong game state: %s", game.State),
			}
		}
	})
}

func (gm *GameManager) handleLobby(ctx context.Context, game *qg.Game, q qg.Querier) (*models.GameUpdate, error) {
	status, err := q.UpdateGameStatus(ctx, qg.UpdateGameStatusParams{
		ID:          game.ID,
		State:       qg.GameStateSubmitting,
		NextStateIn: db.ToPgInterval(gm.config.SubmittingState),
	})

	if err != nil {
		return nil, common.ServiceError{
			Code:    common.GetDbErrorCode(err),
			Message: fmt.Sprintf("wrong game state: %s", game.State),
		}
	}

	upd := statusRowToGameUpdate(status)
	return &upd, nil
}

func (gm *GameManager) handleSubmitting(ctx context.Context, game *qg.Game, q qg.Querier) (*models.GameUpdate, error) {
	status, err := q.UpdateGameStatus(ctx, qg.UpdateGameStatusParams{
		ID:          game.ID,
		State:       qg.GameStateWatching,
		NextStateIn: db.ToPgInterval(gm.config.WatchingState),
	})

	if err != nil {
		return nil, common.ServiceError{
			Code:    common.GetDbErrorCode(err),
			Message: fmt.Sprintf("wrong game state: %s", game.State),
		}
	}

	upd := statusRowToGameUpdate(status)
	return &upd, nil
}

func (gm *GameManager) handleWathching(ctx context.Context, game *qg.Game, q qg.Querier) (*models.GameUpdate, error) {
	completeGame, err := q.SetCompletedStatus(ctx, qg.SetCompletedStatusParams{
		ID:    game.ID,
		State: qg.GameStateCompleted,
	})

	if err != nil {
		return nil, common.ServiceError{
			Code:    common.GetDbErrorCode(err),
			Message: fmt.Sprintf("wrong game state: %s", game.State),
		}
	}

	upd := gameToGameUpdate(completeGame)
	return &upd, nil
}

func statusRowToGameUpdate(row qg.UpdateGameStatusRow) models.GameUpdate {
	return models.GameUpdate{
		GameID:            row.ID,
		MsgType:           models.GameUpdateMsg,
		State:             row.State,
		StateChangedAt:    row.StateChangedAt,
		NextStateChangeAt: row.NextStateChangeAt,
		RamaningTimeMs:    row.RemainingMs,
	}
}

func gameToGameUpdate(game qg.Game) models.GameUpdate {
	return models.GameUpdate{
		GameID:            game.ID,
		MsgType:           models.GameUpdateMsg,
		State:             game.State,
		StateChangedAt:    game.StateChangedAt,
		NextStateChangeAt: game.NextStateChangeAt,
		RamaningTimeMs:    0,
	}
}

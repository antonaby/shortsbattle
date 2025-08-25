package services

import (
	"context"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
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

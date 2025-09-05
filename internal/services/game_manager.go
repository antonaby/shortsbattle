package services

import (
	"context"
	"fmt"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
)

type GameManager struct {
	config GameConfig
	txm    db.TxManager
	redis  *redis.Client
	vs     *VideoService
}

func NewGameManager(txm db.TxManager, redis *redis.Client, vs *VideoService, config GameConfig) *GameManager {
	return &GameManager{
		txm:    txm,
		redis:  redis,
		vs:     vs,
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

func (gm *GameManager) SubmitExistingVideo(ctx context.Context, gameId, videoId, playerId int64, roundN int32) (*qg.Game, *qg.Video, error) {
	return db.WithTxValue2(ctx, gm.txm, func(ctx context.Context, tx pgx.Tx) (*qg.Game, *qg.Video, error) {
		q := gm.txm.Querier(tx)
		game, err := gm.findGameForPlayer(ctx, q, gameId, playerId, []string{string(qg.GameStateLobby), string(qg.GameStateSubmitting)})
		if err != nil {
			return nil, nil, err
		}

		err = gm.addVideoToGame(ctx, q, gameId, videoId, playerId, roundN)
		if err != nil {
			return nil, nil, err
		}

		video, err := q.GetVideo(ctx, qg.GetVideoParams{
			ID: videoId,
		})

		if err != nil {
			return nil, nil, common.ServiceError{
				Code:    common.GetDbErrorCode(err),
				Message: "failed to add video to player",
				Cause:   err,
			}
		}

		return game, &video, nil
	})
}

func (gm *GameManager) SubmitNewVideo(ctx context.Context, gameId int64, videoUrl string, playerId int64, roundN int32) (*qg.Game, *qg.Video, error) {
	oembed, err := fetchOEmbed(ctx, videoUrl, "")
	if err != nil {
		return nil, nil, common.ServiceError{
			Code:    common.ErrorOEmbedFailed,
			Message: "can't get oembed data",
			Cause:   err,
		}
	}

	return db.WithTxValue2(ctx, gm.txm, func(ctx context.Context, tx pgx.Tx) (*qg.Game, *qg.Video, error) {
		q := gm.txm.Querier(tx)
		game, err := gm.findGameForPlayer(ctx, q, gameId, playerId, []string{string(qg.GameStateLobby), string(qg.GameStateSubmitting)})
		if err != nil {
			return nil, nil, err
		}

		video, err := q.AddVideoToPlayer(ctx, qg.AddVideoToPlayerParams{
			PPlayerID: playerId,
			PVideoUrl: videoUrl,
			POembed:   oembed,
		})

		if err != nil {
			return nil, nil, common.ServiceError{
				Code:    common.GetDbErrorCode(err),
				Message: "failed to add video to player",
				Cause:   err,
			}
		}

		err = gm.addVideoToGame(ctx, q, gameId, video.ID, playerId, roundN)
		if err != nil {
			return nil, nil, err
		}

		return game, &video, nil
	})
}

func (gm *GameManager) findGameForPlayer(ctx context.Context, q qg.Querier, gameId, playerId int64, states []string) (*qg.Game, error) {
	game, err := q.FindGameWithPlayer(ctx, qg.FindGameWithPlayerParams{
		GameID:   gameId,
		PlayerID: playerId,
		States:   states,
	})

	if err != nil {
		if db.IsNoRows(err) {
			return nil, common.ServiceError{
				Code:    common.ErrorForbidden,
				Message: "player not in the game",
				Cause:   err,
			}
		}

		return nil, common.ServiceError{
			Code:    common.GetDbErrorCode(err),
			Message: "failed to fetch game",
			Cause:   err,
		}
	}

	return &game, nil
}

func (gm *GameManager) addVideoToGame(ctx context.Context, q qg.Querier, gameId, videoId, playerId int64, roundN int32) error {
	_, err := q.UpsertGameVideoIfOwned(ctx, qg.UpsertGameVideoIfOwnedParams{
		GameID:   gameId,
		RoundN:   roundN,
		PlayerID: playerId,
		VideoID:  videoId,
	})

	if err != nil {
		return common.ServiceError{
			Code:    common.GetDbErrorCode(err),
			Message: "failed to add video to game",
			Cause:   err,
		}
	}

	return nil
}

func (gm *GameManager) GetVideosToWatch(ctx context.Context, gameId, playerId int64, roundN int32) ([]qg.GetVideosToWatchRow, error) {
	return db.WithTxValue(ctx, gm.txm, func(ctx context.Context, tx pgx.Tx) ([]qg.GetVideosToWatchRow, error) {
		q := gm.txm.Querier(tx)
		_, err := gm.findGameForPlayer(ctx, q, gameId, playerId, []string{string(qg.GameStateWatching)})
		if err != nil {
			return nil, err
		}

		videos, err := q.GetVideosToWatch(ctx, qg.GetVideosToWatchParams{
			GameID: gameId,
			RoundN: roundN,
		})

		if err != nil {
			return nil, common.ServiceError{
				Code:    common.GetDbErrorCode(err),
				Message: "failed to get videos for game",
				Cause:   err,
			}
		}

		if len(videos) == 0 {
			videos = []qg.GetVideosToWatchRow{}
		}

		return videos, nil
	})
}

func (gm *GameManager) VoteForVideo(ctx context.Context, params qg.VoteForVideoParams) (*qg.GameVote, error) {
	return db.WithTxValue(ctx, gm.txm, func(ctx context.Context, tx pgx.Tx) (*qg.GameVote, error) {
		q := gm.txm.Querier(tx)

		ok, err := q.PlayerCanVote(ctx, qg.PlayerCanVoteParams{
			GameVideoID: params.GameVideoID,
			PlayerID:    params.PlayerID,
			States:      []string{string(qg.GameStateWatching)},
		})

		if err != nil {
			return nil, common.ServiceError{
				Code:    common.GetDbErrorCode(err),
				Message: "failed to get videos for game",
				Cause:   err,
			}
		}

		if !ok {
			return nil, common.ServiceError{
				Code:    common.ErrorForbidden,
				Message: "player not in the game",
				Cause:   err,
			}
		}

		vote, err := q.VoteForVideo(ctx, params)
		if err != nil {
			return nil, common.ServiceError{
				Code:    common.GetDbErrorCode(err),
				Message: "failed to get videos for game",
				Cause:   err,
			}
		}

		return &vote, nil
	})
}

func (gm *GameManager) GetGameDetailsForPlayer(ctx context.Context, gameId int64, playerId int64) (*models.GameUpdate, error) {
	return db.WithTxValue(ctx, gm.txm, func(ctx context.Context, tx pgx.Tx) (*models.GameUpdate, error) {
		q := gm.txm.Querier(tx)
		game, err := q.GetGameForPlayer(ctx, qg.GetGameForPlayerParams{ID: gameId, PlayerID: playerId})
		if err != nil {
			return nil, common.ServiceError{
				Code:    common.GetDbErrorCode(err),
				Message: "failed to fetch game",
				Cause:   err,
			}
		}

		theme, err := q.GetTheme(ctx, qg.GetThemeParams{ID: game.ThemeID})
		if err != nil {
			return nil, common.ServiceError{
				Code:    common.GetDbErrorCode(err),
				Message: "failed to fetch game theme",
				Cause:   err,
			}
		}

		var round int32
		if game.RoundN.Valid {
			round = game.RoundN.Int32
		}

		return &models.GameUpdate{
			GameID:            game.ID,
			MsgType:           models.GameDetailsMsg,
			State:             game.State,
			RoundN:            round,
			StateChangedAt:    game.StateChangedAt,
			NextStateChangeAt: game.NextStateChangeAt,
			RamaningTimeMs:    game.RemainingMs,
			Theme:             &theme,
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
		RoundN: pgtype.Int4{
			Int32: 1,
			Valid: true,
		},
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
		RoundN:      game.RoundN,
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
	rounds, err := q.GetGameRounds(ctx, qg.GetGameRoundsParams{ID: game.ID})
	if err != nil {
		return nil, common.ServiceError{
			Code:    common.GetDbErrorCode(err),
			Message: fmt.Sprintf("failed to fetch rounds: %s", game.State),
		}
	}

	if int(game.RoundN.Int32) < len(rounds) {
		status, err := q.UpdateGameStatus(ctx, qg.UpdateGameStatusParams{
			ID:          game.ID,
			State:       qg.GameStateSubmitting,
			NextStateIn: db.ToPgInterval(gm.config.SubmittingState),
			RoundN: pgtype.Int4{
				Int32: game.RoundN.Int32 + 1,
				Valid: true,
			},
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

	// TODO: add totoal result
	return &models.GameUpdate{
		MsgType:           models.GameCompleteMsg,
		GameID:            completeGame.ID,
		State:             completeGame.State,
		StateChangedAt:    completeGame.StateChangedAt,
		NextStateChangeAt: completeGame.NextStateChangeAt,
		RamaningTimeMs:    0,
		Result:            &models.GameResult{},
	}, nil
}

func statusRowToGameUpdate(row qg.UpdateGameStatusRow) models.GameUpdate {
	return models.GameUpdate{
		GameID:            row.ID,
		MsgType:           models.GameUpdateMsg,
		State:             row.State,
		RoundN:            row.RoundN.Int32,
		StateChangedAt:    row.StateChangedAt,
		NextStateChangeAt: row.NextStateChangeAt,
		RamaningTimeMs:    row.RemainingMs,
	}
}

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
)

func gmError(code common.ErrorCode, msg string, err error) error {
	return common.ServiceError{
		Code:    code,
		Message: fmt.Sprintf("game manager: %s", msg),
		Cause:   err,
	}
}

func gmDbError(msg string, err error) error {
	return gmError(
		common.GetDbErrorCode(err),
		msg,
		err,
	)
}

func gmGameUpdError(id int64, err error) error {
	return gmDbError(
		fmt.Sprintf("game update failied: game_id=%d", id),
		err,
	)
}

type GameConfig struct {
	MaxPlayers        int32
	MaxLobbyStage     time.Duration
	SubmittingState   time.Duration
	WatchingState     time.Duration
}

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

func (gm *GameManager) JoinGame(ctx context.Context, themeId int64, playerId int64, mode qg.PlayerGameMode) (int64, error) {
	return db.WithTxVQ(ctx, gm.txm, func(ctx context.Context, q qg.Querier) (int64, error) {
		gameId, err := q.JoinGame(ctx, qg.JoinGameParams{
			ThemeID:           themeId,
			PlayerID:          playerId,
			LobbyStage:        qg.GameStageLobby,
			MaxPlayers:        gm.config.MaxPlayers,
			Mode:              mode,
		})

		if err != nil {
			if db.IsClass23(err) {
				return 0, gmError(common.ErrorDbNotFound, "theme not found", err)
			}

			return 0, gmDbError("filed to join game", err)
		}

		return gameId, nil
	})
}

// TODO: add endpoint
func (gm *GameManager) UpdateGameMode(ctx context.Context, gameId, playerId int64, mode qg.PlayerGameMode) (*qg.GamePlayer, error) {
	return db.WithTxVQ(ctx, gm.txm, func(ctx context.Context, q qg.Querier) (*qg.GamePlayer, error) {
		stages := []string{string(qg.GameStageLobby), string(qg.GameStageLobbyFull), string(qg.GameStageSubmit)}
		_, err := gm.findGame(ctx, q, gameId, playerId, stages)
		if err != nil {
			return nil, err
		}

		gamePlayer, err := q.UpdateGameMode(ctx, qg.UpdateGameModeParams{
			GameID:   gameId,
			PlayerID: playerId,
			Mode:     mode,
		})

		if err != nil {
			return nil, gmDbError("failed to update player mode", err)
		}

		return &gamePlayer, nil
	})
}

func (gm *GameManager) SubmitExistingVideo(ctx context.Context, gameId, videoId, playerId int64, roundN int32) (*qg.GetGameAndPlayerLockRow, *qg.GameVideo, error) {
	return db.WithTxVQ2(ctx, gm.txm, func(ctx context.Context, q qg.Querier) (*qg.GetGameAndPlayerLockRow, *qg.GameVideo, error) {
		game, err := gm.findGame(ctx, q, gameId, playerId, []string{string(qg.GameStageSubmit)})
		if err != nil {
			return nil, nil, err
		}

		if game.Mode != qg.PlayerGameModeSubmitAndVote {
			return nil, nil, gmError(common.ErrorForbidden, "player joined in only watching mode", nil)
		}

		gv, err := gm.addVideoToGame(ctx, q, gameId, videoId, playerId, roundN)
		if err != nil {
			return nil, nil, err
		}

		return game, gv, nil
	})
}

func (gm *GameManager) SubmitNewVideo(ctx context.Context, gameId int64, videoUrl string, playerId int64, roundN int32) (*qg.GetGameAndPlayerLockRow, *qg.GameVideo, error) {
	oembed, err := fetchOEmbed(ctx, videoUrl)
	if err != nil {
		return nil, nil, err
	}

	return db.WithTxVQ2(ctx, gm.txm, func(ctx context.Context, q qg.Querier) (*qg.GetGameAndPlayerLockRow, *qg.GameVideo, error) {
		game, err := gm.findGame(ctx, q, gameId, playerId, []string{string(qg.GameStageSubmit)})
		if err != nil {
			return nil, nil, err
		}

		if game.Mode != qg.PlayerGameModeSubmitAndVote {
			return nil, nil, gmError(common.ErrorForbidden, "player joined in only watching mode", nil)
		}

		video, err := q.AddVideoToPlayer(ctx, qg.AddVideoToPlayerParams{
			PlayerID: playerId,
			VideoUrl: videoUrl,
			Oembed:   oembed,
		})

		if err != nil {
			return nil, nil, gmDbError("failed to add video to player", err)
		}

		gv, err := gm.addVideoToGame(ctx, q, gameId, video.ID, playerId, roundN)
		if err != nil {
			return nil, nil, err
		}

		return game, gv, nil
	})
}

func (gm *GameManager) findGame(ctx context.Context, q qg.Querier, gameId, playerId int64, states []string) (*qg.GetGameAndPlayerLockRow, error) {
	game, err := q.GetGameAndPlayerLock(ctx, qg.GetGameAndPlayerLockParams{
		GameID:   gameId,
		PlayerID: playerId,
		Stages:   states,
	})

	if err != nil {
		if db.IsNoRows(err) {
			return nil, gmError(common.ErrorForbidden, "player not in the game or wrong game stage", err)
		}

		return nil, gmDbError("failed to fetch game", err)
	}

	return &game, nil
}

func (gm *GameManager) addVideoToGame(ctx context.Context, q qg.Querier, gameId, videoId, playerId int64, roundN int32) (*qg.GameVideo, error) {
	gv, err := q.UpsertGameVideoIfOwned(ctx, qg.UpsertGameVideoIfOwnedParams{
		GameID:   gameId,
		RoundN:   roundN,
		PlayerID: playerId,
		VideoID:  videoId,
	})

	if err != nil {
		return nil, gmDbError("failed to add video to game", err)
	}

	return &gv, nil
}

func (gm *GameManager) GetVideosToWatch(ctx context.Context, gameId, playerId int64, roundN int32) ([]qg.GetVideosToWatchRow, error) {
	return db.WithTxVQ(ctx, gm.txm, func(ctx context.Context, q qg.Querier) ([]qg.GetVideosToWatchRow, error) {
		_, err := gm.findGame(ctx, q, gameId, playerId, []string{string(qg.GameStageWatch)})
		if err != nil {
			return nil, err
		}

		videos, err := q.GetVideosToWatch(ctx, qg.GetVideosToWatchParams{
			GameID:   gameId,
			PlayerID: playerId,
			RoundN:   roundN,
		})

		if err != nil {
			return nil, gmDbError("failed to get videos for game", err)
		}

		if len(videos) == 0 {
			videos = []qg.GetVideosToWatchRow{}
		}

		return videos, nil
	})
}

// TODO: check voting mode
func (gm *GameManager) VoteForVideo(ctx context.Context, gameVideoId, playerId int64, value json.RawMessage) (*qg.GameStatus, *qg.GameVote, error) {
	return db.WithTxVQ2(ctx, gm.txm, func(ctx context.Context, q qg.Querier) (*qg.GameStatus, *qg.GameVote, error) {
		game, err := q.GetGameVideosForVote(ctx, qg.GetGameVideosForVoteParams{
			GameVideoID: gameVideoId,
			PlayerID:    playerId,
			Stages:      []string{string(qg.GameStageWatch)},
		})

		if err != nil {
			if db.IsNoRows(err) {
				return nil, nil, gmError(common.ErrorForbidden, "player not in the game or wrong game stage", err)
			}

			return nil, nil, gmDbError("failed to vote", err)
		}

		vote, err := q.VoteForVideo(ctx, qg.VoteForVideoParams{
			GameVideoID: gameVideoId,
			PlayerID:    playerId,
			Value:       value,
		})

		if err != nil {
			return nil, nil, gmDbError("failed to vote", err)
		}

		return &game, &vote, nil
	})
}

func (gm *GameManager) GetGameDetailsForPlayer(ctx context.Context, gameId, playerId int64) (*models.GameUpdate, error) {
	return db.WithTxVQ(ctx, gm.txm, func(ctx context.Context, q qg.Querier) (*models.GameUpdate, error) {
		game, err := q.GetGameLock(ctx, gameId, playerId)
		if err != nil {
			if db.IsNoRows(err) {
				return nil, gmError(common.ErrorForbidden, "player not in the game", err)
			}

			return nil, gmDbError("failed to fetch game", err)
		}

		theme, err := q.GetTheme(ctx, game.ThemeID)
		if err != nil {
			return nil, gmDbError("failed to fetch game theme", err)
		}

		return &models.GameUpdate{
			GameID:          game.GameID,
			MsgType:         models.GameDetailsMsg,
			Stage:           game.Stage,
			RoundN:          game.RoundN,
			StateChangedAt:  game.StateChangedAt,
			RemainingTimeMs: game.RemainingMs,
			Theme:           &theme,
		}, nil
	})
}

func (gm *GameManager) AdvanceGame(ctx context.Context, gameId int64) (*models.GameUpdate, error) {
	return db.WithTxVQ(ctx, gm.txm, func(ctx context.Context, q qg.Querier) (*models.GameUpdate, error) {
		game, err := q.GetGameStateLock(ctx, gameId)
		if err != nil {
			return nil, gmDbError("fetch failed", err)
		}

		switch game.Stage {
		case qg.GameStageLobby:
			return gm.handleLobby(ctx, game, q)
		// case qg.GameStateSubmitting:
		// 	return gm.handleSubmitting(ctx, &game, q)
		// case qg.GameStateWatching:
		// 	return gm.handleWathching(ctx, &game, q)
		default:
			return nil, gmError(
				common.ErrorWrongGameStage,
				fmt.Sprintf("wrong game stage, game_id=%d", game.GameID),
				nil,
			)
		}
	})
}

// TODO: check only active users
// TODO: add bots if nPlayers less than MaxPlayers
func (gm *GameManager) handleLobby(ctx context.Context, game qg.GetGameStateLockRow, q qg.Querier) (*models.GameUpdate, error) {
	nPlayers, err := q.CountPlayersInGame(ctx, game.GameID)
	if err != nil {
		return nil, gmGameUpdError(game.GameID, err)
	}

	isTimeout := game.PastMs >= gm.config.MaxLobbyStage.Milliseconds()
	if nPlayers >= int64(gm.config.MaxPlayers) || isTimeout {
		status, err := gm.updateGameStage(ctx, q, game.GameID, qg.GameStageLobbyFull, game.RoundN)
		if err != nil {
			return nil, err
		}

		reason := models.ReasonLobbyFull
		if isTimeout {
			reason = models.ReasonLobbyTimeout
		}

		upd := statusToGameUpdate(status, reason, 0)
		return &upd, nil
	}

	return nil, nil
}

func (gm *GameManager) updateGameStage(ctx context.Context, q qg.Querier, gameId int64, stage qg.GameStage, roundN int32) (*qg.GameStatus, error) {
	status, err := q.UpdateGameStatus(ctx, qg.UpdateGameStatusParams{
		GameID: gameId,
		Stage:  stage,
		RoundN: roundN,
	})

	if err != nil {
		return nil, gmDbError("failed to update game stage", err)
	}

	return &status, nil
}

func statusToGameUpdate(status *qg.GameStatus, reason models.StageChangeReason, remainingTimeMs int64) models.GameUpdate {
	return models.GameUpdate{
		GameID:            status.GameID,
		MsgType:           models.GameUpdateMsg,
		Stage:             status.Stage,
		StateChangeReason: &reason,
		RoundN:            status.RoundN,
		StateChangedAt:    status.StateChangedAt,
		RemainingTimeMs:   remainingTimeMs,
	}
}

// func (gm *GameManager) fromLobbyToSubmitting(ctx context.Context, game qg.GetGameStateLockRow, q qg.Querier, reason string) (*models.GameUpdate, error) {
// 	status, err := q.UpdateGameStatus(ctx, qg.UpdateGameStatusParams{
// 		GameID:      game.GameID,
// 		Stage:       qg.GameStageSubmit,
// 		NextStateIn: db.ToPgInterval(gm.config.SubmittingState),
// 		RoundN:      1,
// 	})

// 	if err != nil {
// 		return nil, gmGameUpdError(game.GameID, err)
// 	}

// 	upd := statusRowToGameUpdate(status, reason)
// 	return &upd, nil
// }

// func (gm *GameManager) updateRemainingTime(ctx context.Context, game qg.GetGameStateLockRow, q qg.Querier, remainingMicro int64) (*models.GameUpdate, error) {
// 	nextStageIn := pgtype.Interval{
// 		Microseconds: remainingMicro,
// 		Days:         0,
// 		Months:       0,
// 		Valid:        true,
// 	}

// 	_, err := q.UpdateRemainingTime(ctx, nextStageIn, game.GameID)

// 	if err != nil {
// 		return nil, gmGameUpdError(game.GameID, err)
// 	}

// 	return nil, nil
// }

// func (gm *GameManager) handleSubmitting(ctx context.Context, game *qg.FetchGameAndLockRow, q qg.Querier) (*models.GameUpdate, error) {
// 	videos, err := q.FetchSubmittedVideosByPlayers(ctx, qg.FetchSubmittedVideosByPlayersParams{
// 		GameID: game.ID,
// 		RoundN: game.RoundN.Int32,
// 	})

// 	if err != nil {
// 		return nil, gmGameUpdError(game.ID, err)
// 	}

// 	if len(videos) > 0 {
// 		allSubmitted := true
// 		for _, v := range videos {
// 			if !v.GameVideoID.Valid {
// 				// Some player hasn't submitted video
// 				allSubmitted = false
// 				break
// 			}
// 		}

// 		if allSubmitted {
// 			return gm.fromSubmittingToWatching(ctx, game, q, ReasonSubmitAll)
// 		}
// 	}

// 	if game.RemainingMs > gm.config.MinRemainingBeforeChangeMs {
// 		return nil, nil
// 	}

// 	return gm.fromSubmittingToWatching(ctx, game, q, ReasonSubmitTimeout)
// }

// func (gm *GameManager) fromSubmittingToWatching(ctx context.Context, game *qg.FetchGameAndLockRow, q qg.Querier, reason string) (*models.GameUpdate, error) {
// 	status, err := q.UpdateGameStatus(ctx, qg.UpdateGameStatusParams{
// 		ID:          game.ID,
// 		State:       qg.GameStateWatching,
// 		NextStateIn: db.ToPgInterval(gm.config.WatchingState),
// 		RoundN:      game.RoundN,
// 	})

// 	if err != nil {
// 		return nil, gmGameUpdError(game.ID, err)
// 	}

// 	upd := statusRowToGameUpdate(status, reason)
// 	return &upd, nil
// }

// // TODO: add 30 sec for ad after all players have watched
// func (gm *GameManager) handleWathching(ctx context.Context, game *qg.FetchGameAndLockRow, q qg.Querier) (*models.GameUpdate, error) {
// 	votes, err := q.FetchVotesByPlayers(ctx, qg.FetchVotesByPlayersParams{
// 		GameID: game.ID,
// 		RoundN: game.RoundN.Int32,
// 	})

// 	if err != nil {
// 		return nil, gmGameUpdError(game.ID, err)
// 	}

// 	if len(votes) > 0 {
// 		allVoted := true
// 		for _, v := range votes {
// 			if !v.Value.Valid {
// 				// Some player hasn't voted
// 				allVoted = false
// 				break
// 			}
// 		}

// 		if allVoted {
// 			return gm.fromWatchingToSubmittingOrComplete(ctx, game, q, ReasonWatchAll)
// 		}
// 	}

// 	if game.RemainingMs > gm.config.MinRemainingBeforeChangeMs {
// 		return nil, nil
// 	}

// 	return gm.fromWatchingToSubmittingOrComplete(ctx, game, q, ReasonWatchTimeout)
// }

// func (gm *GameManager) fromWatchingToSubmittingOrComplete(ctx context.Context, game *qg.FetchGameAndLockRow, q qg.Querier, reason string) (*models.GameUpdate, error) {
// 	rounds, err := q.GetGameRounds(ctx, qg.GetGameRoundsParams{ID: game.ID})
// 	if err != nil {
// 		return nil, gmGameUpdError(game.ID, err)
// 	}

// 	if int(game.RoundN.Int32) < len(rounds) {
// 		status, err := q.UpdateGameStatus(ctx, qg.UpdateGameStatusParams{
// 			ID:          game.ID,
// 			State:       qg.GameStateSubmitting,
// 			NextStateIn: db.ToPgInterval(gm.config.SubmittingState),
// 			RoundN: pgtype.Int4{
// 				Int32: game.RoundN.Int32 + 1,
// 				Valid: true,
// 			},
// 		})

// 		if err != nil {
// 			return nil, gmGameUpdError(game.ID, err)
// 		}

// 		upd := statusRowToGameUpdate(status, reason)
// 		return &upd, nil
// 	}

// 	completeGame, err := q.SetCompletedStatus(ctx, qg.SetCompletedStatusParams{
// 		ID:    game.ID,
// 		State: qg.GameStateCompleted,
// 	})

// 	if err != nil {
// 		return nil, gmGameUpdError(game.ID, err)
// 	}

// 	// TODO: add totoal result
// 	return &models.GameUpdate{
// 		MsgType:           models.GameCompleteMsg,
// 		GameID:            completeGame.ID,
// 		State:             completeGame.State,
// 		StateChangedAt:    completeGame.StateChangedAt,
// 		NextStateChangeAt: completeGame.NextStateChangeAt,
// 		RamaningTimeMs:    0,
// 		Result:            &models.GameResult{},
// 	}, nil
// }

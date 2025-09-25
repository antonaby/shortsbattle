package services

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
	MaxPlayers             int32
	MaxLobbyStage          time.Duration
	MaxLobbyFullStage      time.Duration
	MaxSubmitStage         time.Duration
	MaxSubmitCompleteStage time.Duration
	MaxWatchStage          time.Duration
	MaxWatchCompleteStage  time.Duration
}

type GameManager struct {
	config    GameConfig
	txm       db.TxManager
	validator *validator.Validate
}

func NewGameManager(txm db.TxManager, config GameConfig) *GameManager {
	return &GameManager{
		txm:       txm,
		config:    config,
		validator: validator.New(),
	}
}

func (gm *GameManager) JoinGame(ctx context.Context, themeId int64, playerId int64, playerMode qg.PlayerGameMode) (int64, error) {
	return db.WithTxVQ(ctx, gm.txm, func(ctx context.Context, q qg.Querier) (int64, error) {
		gameId, err := q.JoinGame(ctx, qg.JoinGameParams{
			ThemeID:          themeId,
			PlayerID:         playerId,
			LobbyStage:       qg.GameStageLobby,
			NextGameUpdateIn: db.ToPgInterval(gm.config.MaxLobbyStage),
			MaxPlayers:       gm.config.MaxPlayers,
			PlayerMode:       playerMode,
		})

		if err != nil {
			if db.IsClass23(err) {
				return 0, gmError(common.ErrorNotFound, "theme not found", err)
			}

			return 0, gmDbError("filed to join game", err)
		}

		return gameId, nil
	})
}

func (gm *GameManager) UpdateGameMode(ctx context.Context, gameId, playerId int64, mode qg.PlayerGameMode) (*qg.GamePlayer, error) {
	return db.WithTxVQ(ctx, gm.txm, func(ctx context.Context, q qg.Querier) (*qg.GamePlayer, error) {
		stages := []qg.GameStage{qg.GameStageLobby}
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

// TODO: check duplicate videos (the same video can be submitted only once)
func (gm *GameManager) SubmitExistingVideo(
	ctx context.Context, gameId, videoId, playerId int64, roundN int32,
) (*qg.GetGameShareLockRow, *models.GameVideo, error) {
	return db.WithTxVQ2(ctx, gm.txm, func(ctx context.Context, q qg.Querier) (*qg.GetGameShareLockRow, *models.GameVideo, error) {
		game, err := gm.findGame(ctx, q, gameId, playerId, []qg.GameStage{qg.GameStageSubmit, qg.GameStageSubmitComplete})
		if err != nil {
			return nil, nil, err
		}

		if game.PlayerMode != qg.PlayerGameModeSubmitAndVote {
			// TODO: return another error
			return nil, nil, gmError(common.ErrorForbidden, "player joined in only watching mode", nil)
		}

		gv, err := gm.addVideoToGame(ctx, q, gameId, videoId, playerId, roundN)
		if err != nil {
			return nil, nil, err
		}

		video, err := q.GetVideo(ctx, gv.VideoID, playerId)
		if err != nil {
			return nil, nil, gmDbError("failed to get video", err)
		}

		return game, &models.GameVideo{
			ID:          gv.ID,
			GameID:      gv.GameID,
			PlayerID:    gv.PlayerID,
			RoundN:      gv.RoundN,
			SubmittedAt: gv.SubmittedAt,
			Video: models.Video{
				ID:       video.ID,
				VideoUrl: video.VideoUrl,
				OEmbed:   video.Oembed,
				AddedAt:  video.AddedAt,
			},
		}, nil
	})
}

// TODO: check duplicate videos (the same video can be submitted only once)
func (gm *GameManager) SubmitNewVideo(
	ctx context.Context, gameId int64, videoUrl string, playerId int64, roundN int32,
) (*qg.GetGameShareLockRow, *models.GameVideo, error) {
	oembed, err := fetchOEmbed(ctx, videoUrl)
	if err != nil {
		return nil, nil, err
	}

	return db.WithTxVQ2(ctx, gm.txm, func(ctx context.Context, q qg.Querier) (*qg.GetGameShareLockRow, *models.GameVideo, error) {
		game, err := gm.findGame(ctx, q, gameId, playerId, []qg.GameStage{qg.GameStageSubmit, qg.GameStageSubmitComplete})
		if err != nil {
			return nil, nil, err
		}

		if game.PlayerMode != qg.PlayerGameModeSubmitAndVote {
			return nil, nil, gmError(common.ErrorForbidden, "player joined in only watching mode", nil)
		}

		origVideo, err := q.AddVideoToPlayer(ctx, qg.AddVideoToPlayerParams{
			PlayerID: playerId,
			VideoUrl: videoUrl,
			Oembed:   oembed,
		})

		if err != nil {
			return nil, nil, gmDbError("failed to add video to player", err)
		}

		gv, err := gm.addVideoToGame(ctx, q, gameId, origVideo.ID, playerId, roundN)
		if err != nil {
			return nil, nil, err
		}

		video, err := q.GetVideo(ctx, gv.VideoID, playerId)
		if err != nil {
			return nil, nil, gmDbError("failed to get video", err)
		}

		return game, &models.GameVideo{
			ID:          gv.ID,
			GameID:      gv.GameID,
			PlayerID:    gv.PlayerID,
			RoundN:      gv.RoundN,
			SubmittedAt: gv.SubmittedAt,
			Video: models.Video{
				ID:       video.ID,
				VideoUrl: video.VideoUrl,
				OEmbed:   video.Oembed,
				AddedAt:  video.AddedAt,
			},
		}, nil
	})
}

func (gm *GameManager) findGame(ctx context.Context, q qg.Querier, gameId, playerId int64, stages []qg.GameStage) (*qg.GetGameShareLockRow, error) {
	game, err := q.GetGameShareLock(ctx, gameId, playerId)

	if err != nil {
		if db.IsNoRows(err) {
			return nil, gmError(common.ErrorForbidden, "player not in the game", err)
		}

		return nil, gmDbError("failed to fetch game", err)
	}

	if slices.Contains(stages, game.Stage) {
		return &game, nil
	}

	return nil, gmError(common.ErrorForbidden, "wrong game stage", err)
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

func (gm *GameManager) GetVideosToWatch(ctx context.Context, gameId, playerId int64, roundN int32) ([]models.GameVideo, error) {
	return db.WithTxVQ(ctx, gm.txm, func(ctx context.Context, q qg.Querier) ([]models.GameVideo, error) {
		_, err := gm.findGame(ctx, q, gameId, playerId, []qg.GameStage{qg.GameStageWatch})
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
			return []models.GameVideo{}, nil
		}

		vModels := make([]models.GameVideo, 0, len(videos))
		for _, v := range videos {
			vModels = append(vModels, models.GameVideo{
				ID:          v.GameVideoID,
				GameID:      v.GameID,
				PlayerID:    v.PlayerID,
				RoundN:      v.RoundN,
				SubmittedAt: v.SubmittedAt,
				Video: models.Video{
					ID:       v.VideoID,
					VideoUrl: v.VideoUrl,
					OEmbed:   v.Oembed,
					AddedAt:  v.AddedAt,
				},
			})
		}

		return vModels, nil
	})
}

func (gm *GameManager) VoteForVideo(ctx context.Context, gameVideoId, playerId int64, value json.RawMessage) (*qg.GetGameVideoShareLockRow, *qg.GameVote, error) {
	return db.WithTxVQ2(ctx, gm.txm, func(ctx context.Context, q qg.Querier) (*qg.GetGameVideoShareLockRow, *qg.GameVote, error) {
		game, err := q.GetGameVideoShareLock(ctx, gameVideoId, playerId)
		if err != nil {
			if db.IsNoRows(err) {
				return nil, nil, gmError(common.ErrorForbidden, "player not in the game or player tried to vote for its own video", err)
			}

			return nil, nil, gmDbError("failed to vote", err)
		}

		if !slices.Contains([]qg.GameStage{qg.GameStageWatch, qg.GameStageWatchComplete}, game.Stage) {
			return nil, nil, gmError(common.ErrorForbidden, "wrong game stage", err)
		}

		vote, err := gm.createVote(ctx, q, game, gameVideoId, playerId, value)
		if err != nil {
			return nil, nil, err
		}

		return &game, vote, nil
	})
}

func (gm *GameManager) createVote(
	ctx context.Context, q qg.Querier, game qg.GetGameVideoShareLockRow, 
	gameVideoId, playerId int64, value json.RawMessage) (*qg.GameVote, error) {
	if game.Mode == qg.GameModeLikedislike {
		var vote models.LikeDislikeVote
		err := json.Unmarshal(value, &vote)
		if err != nil {
			return nil, gmError(common.ErrorBadData, "bad vote data", err)
		}

		err = gm.validator.Struct(vote)
		if err != nil {
			return nil, gmError(common.ErrorBadData, "bad vote data", err)
		}

		gameVote, err := q.VoteForVideoLD(ctx, qg.VoteForVideoLDParams{
			GameVideoID: gameVideoId,
			PlayerID:    playerId,
			Value:       string(vote.Value),
		})

		if err != nil {
			return nil, gmDbError("failed to create vote", err)
		}

		return &gameVote, nil
	}

	return nil, gmError(common.ErrorWrongGameMode, "unknown game mode", nil)
}

func (gm *GameManager) GetGameDetailsForPlayer(ctx context.Context, gameId, playerId int64) (*models.GameUpdate, error) {
	return db.WithTxVQ(ctx, gm.txm, func(ctx context.Context, q qg.Querier) (*models.GameUpdate, error) {
		game, err := q.GetGameShareLock(ctx, gameId, playerId)
		if err != nil {
			if db.IsNoRows(err) {
				return nil, gmError(common.ErrorForbidden, "player not in the game", err)
			}

			return nil, gmDbError("failed to fetch game", err)
		}

		theme, err := getTheme(ctx, q, game.ThemeID)
		if err != nil {
			return nil, gmDbError("failed to fetch game theme", err)
		}

		return &models.GameUpdate{
			GameID:         game.GameID,
			PlayerMode:     game.PlayerMode,
			GameMode:       game.Mode,
			MsgType:        models.GameDetailsMsg,
			Stage:          game.Stage,
			RoundN:         game.RoundN,
			StateChangedAt: game.StateChangedAt,
			RemainingMs:    game.RemainingMs,
			Theme:          theme,
		}, nil
	})
}

func (gm *GameManager) GetGameResult(ctx context.Context, gameId, playerId int64) (*models.GameResult, error) {
	return db.WithTxVQ(ctx, gm.txm, func(ctx context.Context, q qg.Querier) (*models.GameResult, error) {
		game, err := gm.findGame(ctx, q, gameId, playerId, []qg.GameStage{qg.GameStageComplete})
		if err != nil {
			return nil, err
		}

		if game.Mode == qg.GameModeLikedislike {
			gameVideos, err := q.GetGameVideoLDResults(ctx, gameId)
			if err != nil {
				return nil, gmDbError("failed to get game videos", err)
			}

			rounds := make(map[int32]models.RoundResult, 0)
			for _, gv := range gameVideos {
				round, ok := rounds[gv.RoundN]
				if !ok {
					round = models.RoundResult{
						Round: models.Round{
							RoundN:      gv.RoundN,
							Title:       gv.RoundTitle,
							Description: gv.RoundDescription.String,
						},
					}
				}

				round.Videos = append(round.Videos, models.VideoResult{
					Video: models.Video{
						ID:       gv.VideoID,
						VideoUrl: gv.VideoUrl,
						OEmbed:   gv.Oembed,
						AddedAt:  gv.AddedAt,
					},
					Author: models.VideoAuthor{
						TgID:     gv.TgID,
						Username: gv.TgUsername,
					},
					Likes:    int(gv.Likes),
					Dislikes: int(gv.Dislikes),
				})

				rounds[gv.RoundN] = round
			}

			roundsRaw := make([]models.RoundResult, 0, len(rounds))
			for _, r := range rounds {
				roundsRaw = append(roundsRaw, r)
			}

			sort.Slice(roundsRaw, func(i, j int) bool {
				return roundsRaw[i].Round.RoundN < roundsRaw[j].Round.RoundN
			})

			return &models.GameResult{
				Rounds: roundsRaw,
			}, nil
		}

		return nil, gmError(common.ErrorWrongGameMode, "unknow game mode", nil)
	})
}

func (gm *GameManager) GetPlayerScores(ctx context.Context, gameId, playerId int64) ([]qg.PlayerResult, error) {
	return db.WithTxVQ(ctx, gm.txm, func(ctx context.Context, q qg.Querier) ([]qg.PlayerResult, error) {
		results, err := q.GetPlayersResult(ctx, gameId, playerId)
		if err != nil {
			return nil, gmDbError("failed to get player stats", err)
		}

		if len(results) == 0 {
			return []qg.PlayerResult{}, nil
		}

		return results, nil
	})
}

func (gm *GameManager) AdvanceGameNow(ctx context.Context, gameId int64) (*models.GameUpdate, error) {
	return db.WithTxVQ(ctx, gm.txm, func(ctx context.Context, q qg.Querier) (*models.GameUpdate, error) {
		pgUUID := pgtype.UUID{Valid: false}
		game, err := q.GetGameLock(ctx, gameId, pgUUID)
		if err != nil {
			if db.IsNoRows(err) {
				return nil, nil
			}

			return nil, gmDbError("failed to fetch game for update", err)
		}

		return gm.advanceGame(ctx, q, game, false)
	})
}

func (gm *GameManager) AdvanceGameAt(ctx context.Context, gameId int64, updateKey uuid.UUID, isTimeout bool) (*models.GameUpdate, error) {
	return db.WithTxVQ(ctx, gm.txm, func(ctx context.Context, q qg.Querier) (*models.GameUpdate, error) {
		pgUUID := pgtype.UUID{Bytes: updateKey, Valid: true}
		game, err := q.GetGameLock(ctx, gameId, pgUUID)
		if err != nil {
			if db.IsNoRows(err) {
				return nil, nil
			}

			return nil, gmDbError("failed to fetch game for update", err)
		}

		return gm.advanceGame(ctx, q, game, isTimeout)
	})
}

// TODO: verify that all errors should be retried (or exclude some of them)
// If it throws an error, asynq may bot need to retry the game advance
func (gm *GameManager) advanceGame(ctx context.Context, q qg.Querier, game qg.GetGameLockRow, isTimeout bool) (*models.GameUpdate, error) {
	switch game.Stage {
	case qg.GameStageLobby:
		return gm.handleLobby(ctx, q, game, isTimeout)
	case qg.GameStageLobbyFull:
		return gm.handleLobbyFull(ctx, q, game, isTimeout)
	case qg.GameStageSubmit:
		return gm.handleSubmit(ctx, q, game, isTimeout)
	case qg.GameStageSubmitComplete:
		return gm.handleSubmitComplete(ctx, q, game, isTimeout)
	case qg.GameStageWatch:
		return gm.handleWatch(ctx, q, game, isTimeout)
	case qg.GameStageWatchComplete:
		return gm.handleWatchComplete(ctx, q, game, isTimeout)
	default:
		return nil, nil
	}
}

// TODO: count only players in submit_and_vote mode
// TODO: check only active users
// TODO: add bots if nPlayers less than MaxPlayers
func (gm *GameManager) handleLobby(ctx context.Context, q qg.Querier, game qg.GetGameLockRow, isTimeout bool) (*models.GameUpdate, error) {
	probablyTimeout := game.PastMs >= gm.config.MaxLobbyStage.Milliseconds() || isTimeout
	if probablyTimeout {
		status, err := gm.updateGameStatus(ctx, q, game.GameID, qg.GameStageLobbyFull, game.RoundN, gm.config.MaxLobbyFullStage)
		if err != nil {
			return nil, err
		}

		upd := statusToGameUpdate(status, models.ReasonLobbyTimeout)
		return &upd, nil
	}

	nPlayers, err := q.CountPlayersInGame(ctx, game.GameID)
	if err != nil {
		return nil, gmGameUpdError(game.GameID, err)
	}

	if nPlayers >= int64(gm.config.MaxPlayers) {
		status, err := gm.updateGameStatus(ctx, q, game.GameID, qg.GameStageLobbyFull, game.RoundN, gm.config.MaxLobbyFullStage)
		if err != nil {
			return nil, err
		}

		upd := statusToGameUpdate(status, models.ReasonLobbyFull)
		return &upd, nil
	}

	return nil, nil
}

func (gm *GameManager) handleLobbyFull(ctx context.Context, q qg.Querier, game qg.GetGameLockRow, isTimeout bool) (*models.GameUpdate, error) {
	probablyTimeout := game.PastMs >= gm.config.MaxLobbyFullStage.Milliseconds() || isTimeout
	if probablyTimeout {
		status, err := gm.updateGameStatus(ctx, q, game.GameID, qg.GameStageSubmit, 1, gm.config.MaxSubmitStage)
		if err != nil {
			return nil, err
		}

		upd := statusToGameUpdate(status, models.ReasonLobbyFullTimeout)
		return &upd, nil
	}

	return nil, nil
}

// TODO: handle users in only watching mode
// TODO: add bots in case some players inactive
func (gm *GameManager) handleSubmit(ctx context.Context, q qg.Querier, game qg.GetGameLockRow, isTimeout bool) (*models.GameUpdate, error) {
	probablyTimeout := game.PastMs >= gm.config.MaxSubmitStage.Milliseconds() || isTimeout
	if probablyTimeout {
		status, err := gm.updateGameStatus(ctx, q, game.GameID, qg.GameStageSubmitComplete, game.RoundN, gm.config.MaxSubmitCompleteStage)
		if err != nil {
			return nil, err
		}

		upd := statusToGameUpdate(status, models.ReasonSubmitTimeout)
		return &upd, nil
	}

	videos, err := q.GetSubmittedVideosByPlayers(ctx, game.GameID, game.RoundN)
	if err != nil {
		return nil, gmGameUpdError(game.GameID, err)
	}

	allPlayersSubmitted := len(videos) > 0
	for _, v := range videos {
		if !v.GameVideoID.Valid {
			allPlayersSubmitted = false
			break
		}
	}

	if allPlayersSubmitted {
		status, err := gm.updateGameStatus(ctx, q, game.GameID, qg.GameStageSubmitComplete, game.RoundN, gm.config.MaxSubmitCompleteStage)
		if err != nil {
			return nil, err
		}

		upd := statusToGameUpdate(status, models.ReasonSubmitAll)
		return &upd, nil
	}

	return nil, nil
}

func (gm *GameManager) handleSubmitComplete(ctx context.Context, q qg.Querier, game qg.GetGameLockRow, isTimeout bool) (*models.GameUpdate, error) {
	probablyTimeout := game.PastMs >= gm.config.MaxSubmitCompleteStage.Milliseconds() || isTimeout
	if probablyTimeout {
		status, err := gm.updateGameStatus(ctx, q, game.GameID, qg.GameStageWatch, game.RoundN, gm.config.MaxWatchStage)
		if err != nil {
			return nil, err
		}

		upd := statusToGameUpdate(status, models.ReasinSubmitCompleteTimeout)
		return &upd, nil
	}

	return nil, nil
}

// TODO: increase time as players vote
func (gm *GameManager) handleWatch(ctx context.Context, q qg.Querier, game qg.GetGameLockRow, isTimeout bool) (*models.GameUpdate, error) {
	probablyTimeout := game.PastMs >= gm.config.MaxWatchStage.Milliseconds() || isTimeout
	if probablyTimeout {
		status, err := gm.updateGameStatus(ctx, q, game.GameID, qg.GameStageWatchComplete, game.RoundN, gm.config.MaxWatchCompleteStage)
		if err != nil {
			return nil, err
		}

		upd := statusToGameUpdate(status, models.ReasonWatchTimeout)
		return &upd, nil
	}

	votes, err := q.GetVotesForRound(ctx, game.GameID, game.RoundN)
	if err != nil {
		return nil, gmGameUpdError(game.GameID, err)
	}

	allPlayersVoted := len(votes) > 0
	for _, v := range votes {
		if !v.VotedAt.Valid {
			allPlayersVoted = false
			break
		}
	}

	if allPlayersVoted {
		status, err := gm.updateGameStatus(ctx, q, game.GameID, qg.GameStageWatchComplete, game.RoundN, gm.config.MaxWatchCompleteStage)
		if err != nil {
			return nil, err
		}

		upd := statusToGameUpdate(status, models.ReasonWatchAll)
		return &upd, nil
	}

	return nil, nil
}

func (gm *GameManager) handleWatchComplete(ctx context.Context, q qg.Querier, game qg.GetGameLockRow, isTimeout bool) (*models.GameUpdate, error) {
	probablyTimeout := game.PastMs >= gm.config.MaxWatchCompleteStage.Milliseconds() || isTimeout
	if probablyTimeout {
		rounds, err := q.GetGameRounds(ctx, game.GameID)
		if err != nil {
			return nil, gmGameUpdError(game.GameID, err)
		}

		if int(game.RoundN) < len(rounds) {
			status, err := gm.updateGameStatus(ctx, q, game.GameID, qg.GameStageSubmit, game.RoundN+1, gm.config.MaxSubmitStage)
			if err != nil {
				return nil, err
			}

			upd := statusToGameUpdate(status, models.ReasonWatchCompleteNextRound)
			return &upd, nil
		}

		// TODO: complete game itself also
		status, err := q.UpdateGameStatusComplete(ctx, qg.GameStageComplete, game.GameID)
		if err != nil {
			return nil, gmDbError("failed to update game stage", err)
		}

		err = gm.finalizeGame(ctx, q, game)
		if err != nil {
			return nil, err
		}

		reason := models.ReasonWatchCompleteGameComplete
		upd := models.GameUpdate{
			GameID:            status.GameID,
			MsgType:           models.GameUpdateMsg,
			GameMode:          status.Mode,
			Stage:             status.Stage,
			StateChangeReason: &reason,
			RoundN:            status.RoundN,
			StateChangedAt:    status.StateChangedAt,
		}
		return &upd, nil
	}

	return nil, nil
}

func (gm *GameManager) updateGameStatus(
	ctx context.Context, q qg.Querier, gameId int64,
	stage qg.GameStage, roundN int32, nextGameChangeIn time.Duration) (*qg.UpdateGameStatusRow, error) {
	status, err := q.UpdateGameStatus(ctx, qg.UpdateGameStatusParams{
		GameID:           gameId,
		Stage:            stage,
		RoundN:           roundN,
		NextGameUpdateIn: db.ToPgInterval(nextGameChangeIn),
	})

	if err != nil {
		return nil, gmDbError("failed to update game stage", err)
	}

	return &status, nil
}

func (gm *GameManager) finalizeGame(ctx context.Context, q qg.Querier, game qg.GetGameLockRow) error {
	if game.Mode == qg.GameModeLikedislike {
		votes, err := q.GetLDVotes(ctx, game.GameID)
		if err != nil {
			return gmDbError("failed to get votes", err)
		}

		if len(votes) == 0 {
			return nil
		}

		likesDislikes, err := calculateLikesAndDislikes(votes)
		if err != nil {
			return err
		}

		err = createGameLDResults(ctx, q, likesDislikes, game)
		if err != nil {
			return err
		}

		err = createPlayerLDResults(ctx, q, likesDislikes, game)
		if err != nil {
			return err
		}

		return nil
	}

	return nil
}

func createGameLDResults(ctx context.Context, q qg.Querier, likesDislikes []models.LikeDislikeVideoResult, game qg.GetGameLockRow) error {
	for _, ld := range likesDislikes {
		_, err := q.CreateGameVideoResult(ctx, qg.CreateGameVideoResultParams{
			GameID:      game.GameID,
			GameVideoID: ld.GameVideoID,
			Likes:       int32(ld.Likes),
			Dislikes:    int32(ld.Dislikes),
		})

		if err != nil {
			return gmDbError("failed to create game video result", err)
		}
	}

	return nil
}

func createPlayerLDResults(ctx context.Context, q qg.Querier, likesDislikes []models.LikeDislikeVideoResult, game qg.GetGameLockRow) error {
	playerLDCount := make(map[int64]models.LikeDislikeCount)
	playerResults := make(map[int64]qg.PlayerResult)

	for _, ld := range likesDislikes {
		player, ok := playerResults[ld.AuthorID]
		if !ok {
			player = qg.PlayerResult{
				PlayerID: ld.AuthorID,
				GameID:   game.GameID,
			}
		}

		player.Points += int64(ld.Likes)
		playerResults[ld.AuthorID] = player

		ldCount, ok := playerLDCount[ld.AuthorID]
		if !ok {
			ldCount = models.LikeDislikeCount{}
		}

		ldCount.Likes += ld.Likes
		ldCount.Dislikes += ld.Dislikes

		playerLDCount[ld.AuthorID] = ldCount
	}

	for playerId, result := range playerResults {
		ldCount, ok := playerLDCount[playerId]
		var count models.LikeDislikeCount
		if ok {
			count = ldCount
		}

		ldCountRaw, err := json.Marshal(count)
		if err != nil {
			return gmError(common.ErrorMarshal, "bad result data", err)
		}

		_, err = q.CreatePlayerResult(ctx, qg.CreatePlayerResultParams{
			PlayerID: result.PlayerID,
			GameID:   result.GameID,
			Points:   result.Points,
			Result:   ldCountRaw,
		})

		if err != nil {
			return gmDbError("failed to create player result", err)
		}
	}

	return nil
}

func calculateLikesAndDislikes(rawVotes []qg.GetLDVotesRow) ([]models.LikeDislikeVideoResult, error) {
	votes := make(map[int64]models.LikeDislikeVideoResult)

	for _, rawVote := range rawVotes {
		if !rawVote.VotedAt.Valid {
			continue
		}

		res, ok := votes[rawVote.GameVideoID]
		if !ok {
			res = models.LikeDislikeVideoResult{
				GameVideoID: rawVote.GameVideoID,
				RoundN:      rawVote.RoundN,
				AuthorID:    rawVote.AuthorID,
			}
		}

		switch models.LikeDislikeVoteValue(rawVote.Value) {
		case models.LikeValue:
			res.Likes += 1
		case models.DislikeValue:
			res.Dislikes += 1
		}

		votes[rawVote.GameVideoID] = res
	}

	if len(votes) == 0 {
		return []models.LikeDislikeVideoResult{}, nil
	}

	results := make([]models.LikeDislikeVideoResult, 0, len(votes))
	for _, v := range votes {
		results = append(results, v)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].RoundN == results[j].RoundN {
			return results[i].Likes > results[j].Likes
		}
		return results[i].RoundN < results[j].RoundN
	})

	return results, nil
}

func statusToGameUpdate(status *qg.UpdateGameStatusRow, reason models.StageChangeReason) models.GameUpdate {
	return models.GameUpdate{
		GameID:            status.GameID,
		GameMode:          status.Mode,
		MsgType:           models.GameUpdateMsg,
		Stage:             status.Stage,
		StateChangeReason: &reason,
		RoundN:            status.RoundN,
		StateChangedAt:    status.StateChangedAt,
		RemainingMs:       status.RemainingMs,
	}
}

package services

import (
	"context"
	"errors"

	"fmt"

	"sync"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/antonaby/shortsbattle/game-server/internal/utils"
	"github.com/jackc/pgx/v5"
)

type GMErrorCode int

const (
	GMErrUnknown GMErrorCode = iota
	GMErrNotFound
	GMErrConstraintViolation
	GMErrDbError
	GMErrTimeout
)

type GameManagerError struct {
	Code    GMErrorCode
	Message string
	Cause   error
}

func (e GameManagerError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("Error %d: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

func (e GameManagerError) Unwrap() error {
	return e.Cause
}

type GameManager struct {
	txm    db.TxManager
	mu     sync.RWMutex
	rounds map[int64]*Round
}

func NewGameManager(txm db.TxManager) *GameManager {
	return &GameManager{
		txm:    txm,
		rounds: make(map[int64]*Round),
	}
}

func (g *GameManager) CreateGame(ctx context.Context, themeId int64) (*db.Game, error) {
	game, err := g.createGame(ctx, themeId)

	if err != nil {
		return nil, err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	round := NewRound(g.txm, *game, g.getDefaultRoundConfig())
	g.rounds[game.ID] = round

	go g.watchRound(round)
	go round.Run()

	return game, nil
}

func (g *GameManager) createGame(ctx context.Context, themeId int64) (*db.Game, error) {
	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*db.Game, error) {
		q := g.txm.Querier(tx)
		game, err := q.CreateGame(ctx, db.CreateGameParams{ThemeID: themeId, Status: db.GameStatusCreated})
		if err != nil {
			if utils.IsClass23(err) {
				return nil, GameManagerError{
					Code:    GMErrConstraintViolation,
					Message: "can't create game",
					Cause:   err,
				}
			}

			return nil, GameManagerError{
				Code:    GMErrDbError,
				Message: "can't create game",
				Cause:   err,
			}
		}

		return &game, nil
	})
}

func (g *GameManager) getDefaultRoundConfig() RoundConfig {
	return RoundConfig{
		MinPlayers:        1,
		MaxPlayers:        8,
		LobbyTimeout:      30 * time.Second,
		SubmittingTimeout: 30 * time.Second,
		VotingTimeout:     30 * time.Second,
		AddPlayerTimeout:  3 * time.Second,
		AddVideoTimeout:   3 * time.Second,
		AddVoteTimeout:    3 * time.Second,
	}
}

func (g *GameManager) watchRound(round *Round) {
	for upd := range round.StatusUpdate {
		if upd.Error != nil {
			break
		}
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	delete(g.rounds, round.Game.ID)
}

func (g *GameManager) GetGame(ctx context.Context, gameId int64) (*models.GameDetails, error) {
	game, err := db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*db.Game, error) {
		q := g.txm.Querier(tx)
		game, err := q.GetGame(ctx, db.GetGameParams{ID: gameId})

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, GameManagerError{
					Code:    GMErrNotFound,
					Message: "game not found",
				}
			}

			return nil, GameManagerError{
				Code:    GMErrDbError,
				Message: "can't get game",
				Cause:   err,
			}
		}

		return &game, nil
	})

	if err != nil {
		return nil, err
	}

	details := &models.GameDetails{
		ID:                 game.ID,
		ThemeID:            game.ThemeID,
		Status:             game.Status,
		CreatedAt:          game.CreatedAt,
		StageTimeRemaining: 0,
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	round, err := g.getRound(game.ID)
	if err == nil {
		details.StageTimeRemaining = round.StageCountdown.Remaining()
	}

	return details, nil
}

func (g *GameManager) getRound(gameId int64) (*Round, error) {
	round, ok := g.rounds[gameId]
	if !ok {
		return nil, GameManagerError{
			Code:    GMErrNotFound,
			Message: "round not found",
		}
	}

	return round, nil
}

func (g *GameManager) AddPlayer(gameId int64, playerId int64) error {
	round, err := g.getRound(gameId)
	if err != nil {
		return err
	}

	timer := time.NewTimer(round.Config.AddPlayerTimeout)
	defer timer.Stop()

	response := make(chan error, 1)
	round.PlayerJoin <- PlayerJoin{GameID: gameId, PlayerID: playerId, Response: response}

	select {
	case err := <-response:
		return err
	case <-timer.C:
		return GameManagerError{
			Code:    GMErrTimeout,
			Message: "timeout adding player",
		}
	}
}

// func (g *GameManager) SubmitVideo(ctx context.Context, params db.CreateVideoParams) (*db.Video, error) {
// 	round, err := g.getRound(params.GameID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	video, err := db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*db.Video, error) {
// 		q := g.txm.Querier(tx)
// 		err := g.checkPlayerInGameAndGameStatus(ctx, params.GameID, params.PlayerID, db.GameStatusLobby, q)
// 		if err != nil {
// 			return nil, err
// 		}

// 		video, err := q.CreateVideo(ctx, params)
// 		if err != nil {
// 			return nil, GameManagerError{
// 				Code:    CodeDbError,
// 				Message: "can't create video",
// 				Cause:   err,
// 			}
// 		}

// 		return &video, nil
// 	})

// 	if err != nil {
// 		return nil, err
// 	}

// 	err = g.submitWithTimeout(
// 		round.Config.AddVideoTimeout,
// 		func(response chan error) {
// 			round.VideoSubmission <- VideoSubmission{Video: *video, Response: response}
// 		},
// 		"can't add video",
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return video, nil
// }

// func (g *GameManager) SubmitVote(ctx context.Context, params db.CreateVoteParams) (*db.Vote, error) {
// 	round, err := g.getRound(params.GameID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	vote, err := db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*db.Vote, error) {
// 		q := g.txm.Querier(tx)
// 		err := g.checkPlayerInGameAndGameStatus(ctx, params.GameID, params.VoterID, db.GameStatusVoting, q)
// 		if err != nil {
// 			return nil, err
// 		}

// 		vote, err := q.CreateVote(ctx, params)
// 		if err != nil {
// 			return nil, GameManagerError{
// 				Code:    CodeDbError,
// 				Message: "can't create vote",
// 				Cause:   err,
// 			}
// 		}

// 		return &vote, nil
// 	})

// 	if err != nil {
// 		return nil, err
// 	}

// 	err = g.submitWithTimeout(
// 		round.Config.AddVoteTimeout,
// 		func(response chan error) {
// 			round.VoteSubmission <- VoteSubmission{Vote: *vote, Response: response}
// 		},
// 		"can't add vote",
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return vote, nil
// }

// func (g *GameManager) checkPlayerInGameAndGameStatus(ctx context.Context, gameId int64, playerId int64, status db.GameStatus, q db.Querier) error {
// 	ok, err := q.IsPlayerInGame(ctx, db.IsPlayerInGameParams{
// 		GameID:   gameId,
// 		PlayerID: playerId,
// 	})

// 	if err != nil {
// 		return GameManagerError{
// 			Code:    getErrorCode(err),
// 			Message: "can't find player in game",
// 			Cause:   err,
// 		}
// 	}

// 	if !ok {
// 		return GameManagerError{
// 			Code:    CodeNotFound,
// 			Message: "player not in game",
// 		}
// 	}

// 	game, err := q.GetGame(ctx, db.GetGameParams{ID: gameId})
// 	if err != nil {
// 		return GameManagerError{
// 			Code:    getErrorCode(err),
// 			Message: "can't fetch game",
// 			Cause:   err,
// 		}
// 	}

// 	if game.Status != status {
// 		return GameManagerError{
// 			Code:    CodeWrongGameStatus,
// 			Message: "game completed or not started",
// 		}
// 	}

// 	return nil
// }

// func (g *GameManager) GetPlayersInGame(ctx context.Context, gameId int64) ([]db.GetPlayersInGameRow, error) {
// 	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) ([]db.GetPlayersInGameRow, error) {
// 		q := g.txm.Querier(tx)
// 		return q.GetPlayersInGame(ctx, db.GetPlayersInGameParams{GameID: gameId})
// 	})
// }

// func (g *GameManager) GetGames(ctx context.Context) ([]db.Game, error) {
// 	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) ([]db.Game, error) {
// 		q := g.txm.Querier(tx)
// 		return q.GetAllGames(ctx)
// 	})
// }

// func getErrorCode(err error) GameManagerErrorCode {
// 	if errors.Is(err, pgx.ErrNoRows) {
// 		return CodeNotFound
// 	}

// 	return CodeDbError
// }

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
	GMErrCanceled
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
		MaxPlayers:        3,
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
		Players:            []db.Player{},
		Videos:             []db.Video{},
		Votes:              []db.Vote{},
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	round, err := g.getRound(game.ID)
	if err == nil {
		details.StageTimeRemaining = round.StageCountdown.Remaining()
		details.Players = round.Players
		details.Videos = round.Videos
		details.Votes = round.Votes
	}
	// TODO: load players, videos and votes from db

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

func (g *GameManager) AddPlayer(ctx context.Context, params db.AddPlayerToGameParams) error {
	round, err := g.getRound(params.GameID)
	if err != nil {
		return err
	}

	timer := time.NewTimer(round.Config.AddPlayerTimeout)
	defer timer.Stop()

	response := make(chan error, 1)
	round.PlayerJoin <- PlayerJoin{
		Player:   params,
		Response: response,
	}

	select {
	case <-ctx.Done():
		return GameManagerError{
			Code:    GMErrCanceled,
			Message: "context canceled",
			Cause:   ctx.Err(),
		}
	case err := <-response:
		return err
	case <-timer.C:
		return GameManagerError{
			Code:    GMErrTimeout,
			Message: "timeout adding player",
		}
	}
}

func (g *GameManager) SubmitVideo(ctx context.Context, params db.CreateVideoParams) error {
	round, err := g.getRound(params.GameID)
	if err != nil {
		return err
	}

	timer := time.NewTimer(round.Config.AddVideoTimeout)
	defer timer.Stop()

	response := make(chan error, 1)
	round.VideoSubmission <- VideoSubmission{
		Video:    params,
		Response: response,
	}

	select {
	case <-ctx.Done():
		return GameManagerError{
			Code:    GMErrCanceled,
			Message: "context canceled",
			Cause:   ctx.Err(),
		}
	case err := <-response:
		return err
	case <-timer.C:
		return GameManagerError{
			Code:    GMErrTimeout,
			Message: "timeout adding video",
		}
	}
}

// func (g *GameManager) SubmitVideo(ctx context.Context, params db.CreateVideoParams) (*db.Video, error) {
// 	round, err := g.getRound(params.GameID)
// 	if err != nil {
// 		return nil, err
// 	}

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

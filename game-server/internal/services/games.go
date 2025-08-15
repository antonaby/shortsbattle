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
	txm   db.TxManager
	mu    sync.RWMutex
	games map[int64]*GameInstance
}

func NewGameManager(txm db.TxManager) *GameManager {
	return &GameManager{
		txm:   txm,
		games: make(map[int64]*GameInstance),
	}
}

func (g *GameManager) JoinGame(ctx context.Context, themeId int64) (*db.Game, error) {
	game, err := db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*db.Game, error) {
		q := g.txm.Querier(tx)
		games, err := q.FindGamesForTheme(ctx, db.FindGamesForThemeParams{
			ThemeID: themeId,
			Status:  db.GameStatusLobby,
		})

		if err != nil {
			return nil, GameManagerError{
				Code:    GMErrDbError,
				Message: "can't fetch games",
				Cause:   err,
			}
		}

		if len(games) != 0 {
			return g.getGameWithLessPlayers(ctx, games, q)
		}

		return g.createGame(ctx, themeId, q)
	})

	if err != nil {
		return nil, err
	}

	_, ok := g.games[game.ID]
	if !ok {
		g.createGameInstance(*game)
	}

	return game, nil
}

func (g *GameManager) createGame(ctx context.Context, themeId int64, q db.Querier) (*db.Game, error) {
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
}

func (g *GameManager) getGameWithLessPlayers(ctx context.Context, games []db.Game, q db.Querier) (*db.Game, error) {
	var ids []int64
	for _, g := range games {
		ids = append(ids, g.ID)
	}

	playersInGames, err := q.ListGamesWithPlayerCounts(ctx, db.ListGamesWithPlayerCountsParams{
		Ids: ids,
	})

	if err != nil {
		return nil, GameManagerError{
			Code:    GMErrDbError,
			Message: "can't fetch players",
			Cause:   err,
		}
	}

	if len(playersInGames) > 0 {
		players := playersInGames[0]
		for _, g := range games {
			if g.ID == players.ID {
				return &g, nil
			}
		}
	}

	return &games[0], nil
}

func (g *GameManager) createGameInstance(game db.Game) {
	g.mu.Lock()
	defer g.mu.Unlock()

	gi := NewGameInstance(g.txm, game, g.defaultGameConfig())
	g.games[game.ID] = gi

	go g.runGame(gi)
}

func (g *GameManager) defaultGameConfig() GameInstanceConfig {
	return GameInstanceConfig{
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

func (g *GameManager) runGame(gi *GameInstance) {
	err := gi.Run()
	if err != nil {
		// Handle error
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	delete(g.games, gi.Game.ID)
}

func (g *GameManager) GetGame(ctx context.Context, gameId int64) (*models.GameDetails, error) {
	details, err := db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*models.GameDetails, error) {
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

		players, err := q.GetPlayersInGame(ctx, db.GetPlayersInGameParams{GameID: game.ID})
		if err != nil {
			return nil, GameManagerError{
				Code:    GMErrDbError,
				Message: "can't get players",
				Cause:   err,
			}
		}

		if len(players) == 0 {
			players = []db.Player{}
		}

		videos, err := q.GetVideosByGame(ctx, db.GetVideosByGameParams{GameID: game.ID})
		if err != nil {
			return nil, GameManagerError{
				Code:    GMErrDbError,
				Message: "can't get videos",
				Cause:   err,
			}
		}

		if len(videos) == 0 {
			videos = []db.Video{}
		}

		votes, err := q.GetVotesByGame(ctx, db.GetVotesByGameParams{GameID: game.ID})
		if err != nil {
			return nil, GameManagerError{
				Code:    GMErrDbError,
				Message: "can't get votes",
				Cause:   err,
			}
		}

		if len(votes) == 0 {
			votes = []db.Vote{}
		}

		return &models.GameDetails{
			ID:                 game.ID,
			ThemeID:            game.ThemeID,
			Status:             game.Status,
			CreatedAt:          game.CreatedAt,
			StageTimeRemaining: 0,
			Players:            players,
			Videos:             videos,
			Votes:              votes,
		}, nil
	})

	if err != nil {
		return nil, err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	gi, err := g.getGameInstance(details.ID)
	if err == nil {
		details.StageTimeRemaining = gi.StageCountdown.Remaining()
	}

	return details, nil
}

func (g *GameManager) getGameInstance(gameId int64) (*GameInstance, error) {
	gi, ok := g.games[gameId]
	if !ok {
		return nil, GameManagerError{
			Code:    GMErrNotFound,
			Message: "game instance not found",
		}
	}

	return gi, nil
}

func (g *GameManager) AddPlayer(ctx context.Context, params db.AddPlayerToGameParams) error {
	gi, err := g.getGameInstance(params.GameID)
	if err != nil {
		return err
	}

	timer := time.NewTimer(gi.Config.AddPlayerTimeout)
	defer timer.Stop()

	response := make(chan error, 1)
	gi.PlayerJoin <- PlayerJoin{
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
	gi, err := g.getGameInstance(params.GameID)
	if err != nil {
		return err
	}

	timer := time.NewTimer(gi.Config.AddVideoTimeout)
	defer timer.Stop()

	response := make(chan error, 1)
	gi.VideoSubmission <- VideoSubmission{
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

func (g *GameManager) SubmitVote(ctx context.Context, params db.CreateVoteParams) error {
	gi, err := g.getGameInstance(params.GameID)
	if err != nil {
		return err
	}

	timer := time.NewTimer(gi.Config.AddVoteTimeout)
	defer timer.Stop()

	response := make(chan error, 1)
	gi.VoteSubmission <- VoteSubmission{
		Vote:     params,
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
			Message: "timeout adding vote",
		}
	}
}

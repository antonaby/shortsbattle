package services

import (
	"context"
	"database/sql"
	"errors"
	"log"

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
	GMErrGameNotAvailable
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
	txm             db.TxManager
	publisher       GameEventPublisher
	mu              sync.RWMutex
	games           map[int64]*GameInstance
	maxJoinAttempts int
}

func NewGameManager(txm db.TxManager) *GameManager {
	return &GameManager{
		txm:             txm,
		games:           make(map[int64]*GameInstance),
		maxJoinAttempts: 20,
	}
}

func (g *GameManager) SetEventPublisher(publisher GameEventPublisher) {
	g.publisher = publisher
}

func (g *GameManager) JoinGame(ctx context.Context, themeId int64, playerId int64) (*db.Game, error) {
	for attempt := 1; attempt <= g.maxJoinAttempts; attempt++ {
		game, err := db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*db.Game, error) {
			q := g.txm.Querier(tx)
			return g.findGameToJoin(ctx, themeId, q)
		})

		if err != nil {
			return nil, err
		}

		gi := g.createGameInstanceIfNeeded(*game)

		err = g.addPlayer(ctx, gi, db.AddPlayerToGameParams{
			GameID:   game.ID,
			PlayerID: playerId,
		})

		if err == nil {
			return game, nil
		}

		var giErr GameInstanceError
		if errors.As(err, &giErr) {
			if giErr.Code == GIErrConstraintViolation {
				return nil, GameManagerError{
					Code:    GMErrConstraintViolation,
					Message: "can't add player",
					Cause:   giErr,
				}
			}
			if giErr.Code == GIErrDbError {
				return nil, GameManagerError{
					Code:    GMErrDbError,
					Message: "can't add player",
					Cause:   giErr,
				}
			}
		}

		if err := ctx.Err(); err != nil {
			return nil, GameManagerError{
				Code:    GMErrCanceled,
				Message: "context canceled",
				Cause:   err,
			}
		}
	}

	return nil, GameManagerError{
		Code:    GMErrGameNotAvailable,
		Message: "can't find a game after max attempts",
	}
}

// TODO: move to GameInstance
func (g *GameManager) PlayerLeave(ctx context.Context, gameId int64, playerId int64) error {
	return db.WithTx(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) error {
		q := g.txm.Querier(tx)
		err := q.RemovePlayerFromGame(ctx, db.RemovePlayerFromGameParams{
			GameID:   gameId,
			PlayerID: playerId,
		})

		if err != nil {
			return GameManagerError{
				Code:    GMErrDbError,
				Message: "can't remove player from game",
				Cause:   err,
			}
		}

		return nil
	})
}

func (g *GameManager) findGameToJoin(ctx context.Context, themeId int64, q db.Querier) (*db.Game, error) {
	err := q.AcquireAdvisoryXactLock(ctx, db.AcquireAdvisoryXactLockParams{Column1: themeId})
	if err != nil {
		return nil, GameManagerError{
			Code:    GMErrDbError,
			Message: "can't acquire theme lock",
			Cause:   err,
		}
	}

	game, err := q.FindLeastCrowdedGameByThemeAndStatuses(ctx, db.FindLeastCrowdedGameByThemeAndStatusesParams{
		ThemeID: themeId,
		Column2: []string{string(db.GameStatusCreated), string(db.GameStatusLobby)},
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return g.createGame(ctx, themeId, q)
		} else {
			return nil, GameManagerError{
				Code:    GMErrDbError,
				Message: "can't fetch games",
				Cause:   err,
			}
		}
	}

	return &db.Game{
		ID:        game.ID,
		ThemeID:   game.ThemeID,
		Status:    game.Status,
		CreatedAt: game.CreatedAt,
	}, nil
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

func (g *GameManager) createGameInstanceIfNeeded(game db.Game) *GameInstance {
	g.mu.Lock()
	defer g.mu.Unlock()

	gi, ok := g.games[game.ID]
	if ok {
		return gi
	}

	newGI := NewGameInstance(g.txm, g.publisher, game, g.defaultGameConfig())
	g.games[game.ID] = newGI
	go g.runGame(newGI)

	return newGI
}

func (g *GameManager) defaultGameConfig() GameInstanceConfig {
	return GameInstanceConfig{
		MinPlayers:        1,
		MaxPlayers:        3,
		TickerDuration:    2 * time.Second,
		CreatedTimeout:    3 * time.Second,
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
		log.Println(err.Error())
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	delete(g.games, gi.Game.ID)
}

func (g *GameManager) GetGame(ctx context.Context, gameId int64) (*models.GameDetails, error) {
	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*models.GameDetails, error) {
		q := g.txm.Querier(tx)
		return g.getGame(ctx, gameId, q)
	})
}

func (g *GameManager) getGame(ctx context.Context, gameId int64, q db.Querier) (*models.GameDetails, error) {
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

	theme, err := q.GetThemeByID(ctx, db.GetThemeByIDParams{ID: game.ThemeID})
	if err != nil {
		return nil, GameManagerError{
			Code:    GMErrDbError,
			Message: "can't get theme",
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

	details := &models.GameDetails{
		ID:                 game.ID,
		Theme:              theme,
		Status:             game.Status,
		CreatedAt:          game.CreatedAt,
		StageTimeRemaining: 0,
		Players:            players,
		Videos:             videos,
		Votes:              votes,
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	gi, err := g.getGameInstance(details.ID)
	if err == nil {
		details.StageTimeRemaining = gi.StageCountdown.Remaining().Milliseconds()
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

func (g *GameManager) addPlayer(ctx context.Context, gi *GameInstance, params db.AddPlayerToGameParams) error {
	timer := time.NewTimer(gi.Config.AddPlayerTimeout)
	defer timer.Stop()

	response := make(chan error, 1)
	gi.PlayerJoin <- PlayerJoin{
		Player:   params,
		Ctx:      ctx,
		Response: response,
	}

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

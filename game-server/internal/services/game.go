package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type GameServiceErrorCode int

const (
	CodeUnknown GameServiceErrorCode = iota
	CodeNotFound
	CodeDbError
	CodeTimeout
	CodeGameComplete
)

type GameServiceError struct {
	Code    GameServiceErrorCode
	Message string
	Cause   error
}

func (e GameServiceError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("Error %d: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

func (e GameServiceError) Unwrap() error {
	return e.Cause
}

type GameStatusUpdate struct {
	GameID int64
	Status db.GameStatus
	Error  error
}

type PlayerJoinRequest struct {
	Player   db.Player
	Response chan error
}

type VideoSubmission struct {
	Submission db.Video
	Response   error
}

type VoteSubmission struct {
	Vote     db.Vote
	Response error
}

type RoundConfig struct {
	MinPlayers       int
	MaxPlayers       int
	LobbyTimeout     time.Duration
	AddPlayerTimeout time.Duration
}

type Round struct {
	Game            db.Game
	Config          RoundConfig
	Players         []db.Player
	Videos          []db.Video
	Votes           []db.Vote
	PlayerJoin      chan PlayerJoinRequest
	VideoSubmission chan VideoSubmission
	VoteSubmission  chan VoteSubmission
	StatusUpdate    chan GameStatusUpdate
	Ctx             context.Context
	Cancel          context.CancelFunc
}

func NewRound(game db.Game, config RoundConfig) *Round {
	ctx, cancel := context.WithCancel(context.Background())
	return &Round{
		Game:            game,
		Config:          config,
		PlayerJoin:      make(chan PlayerJoinRequest),
		StatusUpdate:    make(chan GameStatusUpdate),
		VideoSubmission: make(chan VideoSubmission),
		VoteSubmission:  make(chan VoteSubmission),
		Ctx:             ctx,
		Cancel:          cancel,
	}
}

func (g *Round) Run() {
	defer g.Cancel()

	g.StatusUpdate <- GameStatusUpdate{
		GameID: g.Game.ID,
		Status: db.GameStatusLobby,
		Error:  nil,
	}

	if !g.LobbyStage() {
		log.Println("Not enought players have joined")
	}

	g.StatusUpdate <- GameStatusUpdate{
		GameID: g.Game.ID,
		Status: db.GameStatusComplete,
		Error:  nil,
	}

	close(g.StatusUpdate)
}

func (g *Round) LobbyStage() bool { // TODO: Add error
	timer := time.NewTimer(g.Config.LobbyTimeout)
	defer timer.Stop()

	for {
		select {
		case <-g.Ctx.Done():
			return false
		case p := <-g.PlayerJoin:
			g.Players = append(g.Players, p.Player)
			p.Response <- nil
			close(p.Response)
			log.Println("Player has been added")
			if len(g.Players) >= g.Config.MaxPlayers {
				return true
			}
		case <-timer.C:
			return len(g.Players) >= g.Config.MinPlayers
		}
	}
}

type GameService struct {
	txm    db.TxManager
	mu     sync.RWMutex
	rounds map[int64]*Round
}

func NewGameService(txm db.TxManager) *GameService {
	return &GameService{
		txm:    txm,
		rounds: make(map[int64]*Round),
	}
}

func (g *GameService) GetGames(ctx context.Context) ([]db.Game, error) {
	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) ([]db.Game, error) {
		q := g.txm.Querier(tx)
		return q.GetAllGames(ctx)
	})
}

func (g *GameService) CreateGame(ctx context.Context) (*db.Game, error) {
	game, err := db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*db.Game, error) {
		q := g.txm.Querier(tx)
		params := db.CreateGameParams{
			Status: db.GameStatusCreated,
			Name:   "Test",
			Description: pgtype.Text{
				String: "Test", Valid: true,
			},
		}
		game, err := q.CreateGame(ctx, params)
		if err != nil {
			return nil, GameServiceError{
				Code:    CodeDbError,
				Message: "can't create game",
				Cause:   err,
			}
		}

		return &game, nil
	})

	if err != nil {
		return nil, err
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	round := NewRound(*game, RoundConfig{
		MinPlayers:       5,
		MaxPlayers:       8,
		LobbyTimeout:     30 * time.Second,
		AddPlayerTimeout: 3 * time.Second,
	})
	g.rounds[game.ID] = round

	go g.watchRound(round)
	go round.Run()

	return game, nil
}

func (g *GameService) watchRound(round *Round) {
	for upd := range round.StatusUpdate {
		err := db.WithTx(context.Background(), g.txm, func(ctx context.Context, tx pgx.Tx) error {
			q := g.txm.Querier(tx)

			var status db.GameStatus
			if upd.Error != nil {
				status = db.GameStatusComplete
			} else {
				status = upd.Status
			}

			return q.UpdateGameStatus(ctx, db.UpdateGameStatusParams{
				Status: status,
				ID:     upd.GameID, // TODO: add message with result
			})
		})

		if err != nil || upd.Error != nil {
			break
		}
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	delete(g.rounds, round.Game.ID)
}

func (g *GameService) AddPlayer(ctx context.Context, gameId int64, playerId int64) error {
	round, ok := g.rounds[gameId]
	if !ok {
		return GameServiceError{
			Code:    CodeNotFound,
			Message: "round not found",
		}
	}

	player, err := db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*db.Player, error) {
		q := g.txm.Querier(tx)
		game, err := q.GetGame(ctx, db.GetGameParams{ID: gameId})
		if err != nil {
			return nil, GameServiceError{
				Code:    CodeDbError,
				Message: "db error fetching game",
				Cause:   err,
			}
		}

		if game.Status == db.GameStatusComplete {
			return nil, GameServiceError{
				Code:    CodeGameComplete,
				Message: "game complete",
			}
		}

		err = q.AddPlayerToGame(ctx, db.AddPlayerToGameParams{
			GameID:   gameId,
			PlayerID: playerId,
		})

		if err != nil {
			return nil, GameServiceError{
				Code:    CodeDbError,
				Message: "db error adding player",
				Cause:   err,
			}
		}

		player, err := q.GetPlayer(ctx, db.GetPlayerParams{ID: playerId})
		if err != nil {
			return nil, GameServiceError{
				Code:    CodeDbError,
				Message: "db error fetching player",
				Cause:   err,
			}
		}

		return &player, nil
	})

	if err != nil {
		return err
	}

	timer := time.NewTimer(round.Config.AddPlayerTimeout)
	defer timer.Stop()

	response := make(chan error, 1)
	round.PlayerJoin <- PlayerJoinRequest{Player: *player, Response: response}

	select {
	case err := <-response:
		if err != nil {
			return GameServiceError{
				Code:    CodeUnknown,
				Message: "can't add player",
				Cause:   err,
			}
		}
	case <-timer.C:
		return GameServiceError{
			Code:    CodeTimeout,
			Message: "can't add player",
		}
	}

	return nil
}

func (g *GameService) SubmitVideo(ctx context.Context, params db.CreateVideoParams) (*db.Video, error) {
	video, err := db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*db.Video, error) {
		q := g.txm.Querier(tx)
		ok, err := q.IsPlayerInGame(ctx, db.IsPlayerInGameParams{
			GameID:   params.GameID,
			PlayerID: params.PlayerID,
		})

		if err != nil {
			return nil, GameServiceError{
				Code:    CodeDbError,
				Message: "can't find player in game",
				Cause:   err,
			}
		}

		if !ok {
			return nil, GameServiceError{
				Code:    CodeNotFound,
				Message: "player not in game",
			}
		}

		game, err := q.GetGame(ctx, db.GetGameParams{ID: params.GameID})
		if err != nil {
			return nil, GameServiceError{
				Code: CodeDbError,
				Message: "can't fetch game",
				Cause: err,
			}
		}

		if game.Status != db.GameStatusLobby {
			return nil, GameServiceError{
				Code: CodeGameComplete,
				Message: "game completed or not started",
			}
		}

		video, err := q.CreateVideo(ctx, params)
		if err != nil {
			return nil, GameServiceError{
				Code:    CodeDbError,
				Message: "can't create video",
				Cause:   err,
			}
		}

		return &video, nil
	})

	if err != nil {
		return nil, err
	}

	return video, nil
}

func (g *GameService) GetPlayersInGame(ctx context.Context, gameId int64) ([]db.GetPlayersInGameRow, error) {
	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) ([]db.GetPlayersInGameRow, error) {
		q := g.txm.Querier(tx)
		return q.GetPlayersInGame(ctx, db.GetPlayersInGameParams{GameID: gameId})
	})
}

func (g *GameService) CreatePlayer(ctx context.Context, params db.CreatePlayerParams) (db.Player, error) {
	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (db.Player, error) {
		q := g.txm.Querier(tx)
		return q.CreatePlayer(ctx, params)
	})
}

func (g *GameService) GetPlayer(ctx context.Context, id int64) (db.Player, error) {
	player, err := db.WithTxValue(
		ctx, g.txm,
		func(ctx context.Context, tx pgx.Tx) (db.Player, error) {
			q := g.txm.Querier(tx)
			return q.GetPlayer(ctx, db.GetPlayerParams{ID: id})
		})

	if err != nil {
		return db.Player{}, err
	}

	return player, err
}

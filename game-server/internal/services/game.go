package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type GameServiceError struct {
	Code    int
	Message string
}

func (e GameServiceError) Error() string {
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
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
	MinPlayers   int
	MaxPlayers   int
	LobbyTimeout time.Duration
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
	txm   db.TxManager
	mu    sync.RWMutex
	games map[int64]*Round
}

func NewGameService(txm db.TxManager) *GameService {
	return &GameService{
		txm:   txm,
		games: make(map[int64]*Round),
	}
}

func (g *GameService) CreateGame(ctx context.Context) (*db.Game, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*db.Game, error) {
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
			return nil, err
		}

		round := NewRound(game, RoundConfig{
			MinPlayers: 5,
			MaxPlayers: 8,
			LobbyTimeout: 30 * time.Second,
		})
		g.games[game.ID] = round

		go g.watchRound(round)
		go round.Run()

		return &game, nil
	})
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
				ID:     upd.GameID,
			})
		})

		if err != nil || upd.Error != nil {
			break
		}
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	delete(g.games, round.Game.ID)
}

func (g *GameService) GetGames(ctx context.Context) ([]db.Game, error) {
	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) ([]db.Game, error) {
		q := g.txm.Querier(tx)
		return q.GetAllGames(ctx)
	})
}

func (g *GameService) AddPlayer(ctx context.Context, gameId int64, playerId int64) error {
	return db.WithTx(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) error {
		q := g.txm.Querier(tx)
		err := q.AddPlayerToGame(ctx, db.AddPlayerToGameParams{
			GameID:   gameId,
			PlayerID: playerId,
		})

		if err != nil {
			return err
		}

		player, err := q.GetPlayer(ctx, playerId)
		if err != nil {
			return err
		}

		game, ok := g.games[gameId]
		if !ok {
			return GameServiceError{
				Code:    404,
				Message: "Game Not Found",
			}
		}

		timer := time.NewTimer(3 * time.Second) // TODO: set proper timeout
		defer timer.Stop()

		response := make(chan error, 1)
		game.PlayerJoin <- PlayerJoinRequest{Player: player, Response: response}

		select {
		case err := <-response:
			if err != nil {
				return err
			}
		case <-timer.C:
			return GameServiceError{
				Code:    500,
				Message: "Can't add a player",
			}
		}

		return nil
	})
}

func (g *GameService) GetPlayersInGame(ctx context.Context, gameId int64) ([]db.GetPlayersInGameRow, error) {
	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) ([]db.GetPlayersInGameRow, error) {
		q := g.txm.Querier(tx)
		return q.GetPlayersInGame(ctx, gameId)
	})
}

func (g *GameService) CreatePlayer(ctx context.Context, params models.CreatePlayerRequest) (db.Player, error) {
	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (db.Player, error) {
		q := g.txm.Querier(tx)
		return q.CreatePlayer(ctx, params.Username)
	})
}

func (g *GameService) GetPlayer(ctx context.Context, id int64) (db.Player, error) {
	player, err := db.WithTxValue(
		ctx, g.txm,
		func(ctx context.Context, tx pgx.Tx) (db.Player, error) {
			q := g.txm.Querier(tx)
			return q.GetPlayer(ctx, id)
		})

	if err != nil {
		return db.Player{}, err
	}

	return player, err
}

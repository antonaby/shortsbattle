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

type GameState struct {
	Game         db.Game
	MaxPlayers   int
	MinPlayers   int
	Players      []db.Player
	Submissions  []db.Submission
	Votes        []db.Vote
	PlayerJoin   chan PlayerJoinRequest
	StatusUpdate chan GameStatusUpdate
	Ctx          context.Context
	Cancel       context.CancelFunc
}

func NewGame(game db.Game, maxPlayers int, minPlayers int) *GameState {
	ctx, cancel := context.WithCancel(context.Background())
	return &GameState{
		Game:         game,
		MaxPlayers:   maxPlayers,
		MinPlayers:   minPlayers,
		PlayerJoin:   make(chan PlayerJoinRequest),
		StatusUpdate: make(chan GameStatusUpdate),
		Ctx:          ctx,
		Cancel:       cancel,
	}
}

func (g *GameState) Run() {
	defer g.Cancel()

	g.StatusUpdate <- GameStatusUpdate{
		GameID: g.Game.ID,
		Status: db.GameStatusLobby,
		Error:  nil,
	}

	if !g.LobbyStage(10 * time.Second) {
		log.Println("Not enought players have joined")
	}

	g.StatusUpdate <- GameStatusUpdate{
		GameID: g.Game.ID,
		Status: db.GameStatusComplete,
		Error:  nil,
	}

	close(g.StatusUpdate)
}

func (g *GameState) LobbyStage(timeout time.Duration) bool { // TODO: Add error
	timer := time.NewTimer(timeout)
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
			if len(g.Players) >= g.MaxPlayers {
				return true
			}
		case <-timer.C:
			return len(g.Players) >= g.MinPlayers
		}
	}
}

type GameService struct {
	txm   db.TxManager
	mu    sync.RWMutex
	games map[int64]*GameState
}

func NewGameService(txm db.TxManager) *GameService {
	return &GameService{
		txm:   txm,
		games: make(map[int64]*GameState),
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

		state := NewGame(game, 8, 5)
		g.games[game.ID] = state

		go func() {
			for upd := range state.StatusUpdate {
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

			delete(g.games, state.Game.ID)
		}()

		go state.Run()

		return &game, nil
	})
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

package services

import (
	"context"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/jackc/pgx/v5"
)


type GameState struct {
	ID int64
	Players []db.Player
	Submissions []db.Submission
	Votes []db.Vote
	Done chan struct{}
	Ctx          context.Context
	Cancel       context.CancelFunc
}

func NewGame(id int64) *GameState {
	ctx, cancel := context.WithCancel(context.Background())
	return &GameState{
		ID: id,
		Done: make(chan struct{}),
		Ctx: ctx,
		Cancel: cancel,
	}
}




type GameService struct {
	txm db.TxManager
}

func NewGameService(txm db.TxManager) *GameService {
	return &GameService{
		txm: txm,
	}
}

func (g GameService) CreateGame(ctx context.Context) (db.Game, error) {
	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (db.Game, error) {
		q := g.txm.Querier(tx)
		return q.CreateGame(ctx, db.GameStatusLobby)
	})
}

func (g GameService) GetGames(ctx context.Context) ([]db.Game, error) {
	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) ([]db.Game, error) {
		q := g.txm.Querier(tx)
		return q.GetAllGames(ctx)
	})
}

func (g GameService) AddPlayer(ctx context.Context, gameId int64, playerId int64) error {
	return db.WithTx(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) error {
		q := g.txm.Querier(tx)
		return q.AddPlayerToGame(ctx, db.AddPlayerToGameParams{
			GameID: gameId,
			PlayerID: playerId,
		})
	})
}

func (g GameService) GetPlayersInGame(ctx context.Context, gameId int64) ([]db.GetPlayersInGameRow, error) {
	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) ([]db.GetPlayersInGameRow, error) {
		q := g.txm.Querier(tx)
		return q.GetPlayersInGame(ctx, gameId)
	})
}

func (g GameService) CreatePlayer(ctx context.Context, params models.CreatePlayerRequest) (db.Player, error) {
	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (db.Player, error) {
		q := g.txm.Querier(tx)
		return q.CreatePlayer(ctx, params.Username)
	})
}

func (g GameService) GetPlayer(ctx context.Context, id int64) (db.Player, error) {
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

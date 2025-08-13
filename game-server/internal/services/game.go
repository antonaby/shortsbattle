package services

import (
	"context"

	"fmt"

	"sync"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/utils"
	"github.com/jackc/pgx/v5"
)

type GameManagerErrorCode int

const (
	GameManagerUnknownErrorCode GameManagerErrorCode = iota
	GameManagerNotFoundErrorCode
	GameManagerConstraintViolationErrroCode
	GameManagerDbErrorCode
)

type GameManagerError struct {
	Code    GameManagerErrorCode
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

	// 	go g.watchRound(round)
	// 	go round.Run()

	return game, nil
}

func (g *GameManager) createGame(ctx context.Context, themeId int64) (*db.Game, error) {
	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*db.Game, error) {
		q := g.txm.Querier(tx)
		game, err := q.CreateGame(ctx, db.CreateGameParams{ThemeID: themeId, Status: db.GameStatusCreated})
		if err != nil {
			if utils.IsClass23(err) {
				return nil, GameManagerError{
					Code:    GameManagerConstraintViolationErrroCode,
					Message: "can't create game",
					Cause:   err,
				}
			}

			return nil, GameManagerError{
				Code:    GameManagerDbErrorCode,
				Message: "can't create game",
				Cause:   err,
			}
		}

		return &game, nil
	})
}

func (g *GameManager) getDefaultRoundConfig() RoundConfig {
	return RoundConfig{
		MinPlayers:       1,
		MaxPlayers:       8,
		LobbyTimeout:     30 * time.Second,
		VotingTimeout:    30 * time.Second,
		AddPlayerTimeout: 3 * time.Second,
		AddVideoTimeout:  3 * time.Second,
		AddVoteTimeout:   3 * time.Second,
	}
}

// func (g *GameManager) watchRound(round *Round) {
// 	for upd := range round.StatusUpdate {
// 		err := db.WithTx(context.Background(), g.txm, func(ctx context.Context, tx pgx.Tx) error {
// 			q := g.txm.Querier(tx)

// 			var status db.GameStatus
// 			if upd.Error != nil {
// 				status = db.GameStatusComplete
// 			} else {
// 				status = upd.Status
// 			}

// 			return q.UpdateGameStatus(ctx, db.UpdateGameStatusParams{
// 				Status: status,
// 				ID:     upd.GameID, // TODO: add message with result
// 			})
// 		})

// 		if err != nil || upd.Error != nil {
// 			break
// 		}
// 	}

// 	g.mu.Lock()
// 	defer g.mu.Unlock()

// 	delete(g.rounds, round.Game.ID)
// }

// func (g *GameManager) AddPlayer(ctx context.Context, gameId int64, playerId int64) error {
// 	round, err := g.getRound(gameId)
// 	if err != nil {
// 		return err
// 	}

// 	player, err := db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*db.Player, error) {
// 		q := g.txm.Querier(tx)
// 		game, err := q.GetGame(ctx, db.GetGameParams{ID: gameId})
// 		if err != nil {
// 			return nil, GameManagerError{
// 				Code:    CodeDbError,
// 				Message: "db error fetching game",
// 				Cause:   err,
// 			}
// 		}

// 		if game.Status != db.GameStatusLobby {
// 			return nil, GameManagerError{
// 				Code:    CodeWrongGameStatus,
// 				Message: "game complete",
// 			}
// 		}

// 		err = q.AddPlayerToGame(ctx, db.AddPlayerToGameParams{
// 			GameID:   gameId,
// 			PlayerID: playerId,
// 		})

// 		if err != nil {
// 			return nil, GameManagerError{
// 				Code:    CodeDbError,
// 				Message: "db error adding player",
// 				Cause:   err,
// 			}
// 		}

// 		player, err := q.GetPlayer(ctx, db.GetPlayerParams{ID: playerId})
// 		if err != nil {
// 			return nil, GameManagerError{
// 				Code:    CodeDbError,
// 				Message: "db error fetching player",
// 				Cause:   err,
// 			}
// 		}

// 		return &player, nil
// 	})

// 	if err != nil {
// 		return err
// 	}

// 	err = g.submitWithTimeout(
// 		round.Config.AddPlayerTimeout,
// 		func(response chan error) {
// 			round.PlayerJoin <- PlayerJoin{Player: *player, Response: response}
// 		},
// 		"can't add player",
// 	)

// 	return err
// }

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

// func (g *GameManager) getRound(gameId int64) (*Round, error) {
// 	round, ok := g.rounds[gameId]
// 	if !ok {
// 		return nil, GameManagerError{
// 			Code:    CodeNotFound,
// 			Message: "round not found",
// 		}
// 	}

// 	return round, nil
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

// func (g *GameManager) submitWithTimeout(timeout time.Duration, sendFunc func(response chan error), errMsg string) error {
// 	timer := time.NewTimer(timeout)
// 	defer timer.Stop()

// 	response := make(chan error, 1)
// 	sendFunc(response)

// 	select {
// 	case err := <-response:
// 		if err != nil {
// 			return GameManagerError{
// 				Code:    CodeUnknown,
// 				Message: errMsg,
// 				Cause:   err,
// 			}
// 		}
// 		return nil
// 	case <-timer.C:
// 		return GameManagerError{
// 			Code:    CodeTimeout,
// 			Message: errMsg,
// 		}
// 	}
// }

// func (g *GameManager) GetPlayersInGame(ctx context.Context, gameId int64) ([]db.GetPlayersInGameRow, error) {
// 	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) ([]db.GetPlayersInGameRow, error) {
// 		q := g.txm.Querier(tx)
// 		return q.GetPlayersInGame(ctx, db.GetPlayersInGameParams{GameID: gameId})
// 	})
// }

// func (g *GameManager) CreatePlayer(ctx context.Context, params db.CreatePlayerParams) (db.Player, error) {
// 	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (db.Player, error) {
// 		q := g.txm.Querier(tx)
// 		return q.CreatePlayer(ctx, params)
// 	})
// }

// func (g *GameManager) GetPlayer(ctx context.Context, id int64) (db.Player, error) {
// 	player, err := db.WithTxValue(
// 		ctx, g.txm,
// 		func(ctx context.Context, tx pgx.Tx) (db.Player, error) {
// 			q := g.txm.Querier(tx)
// 			return q.GetPlayer(ctx, db.GetPlayerParams{ID: id})
// 		})

// 	if err != nil {
// 		return db.Player{}, err
// 	}

// 	return player, err
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

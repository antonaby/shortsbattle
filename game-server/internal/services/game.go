package services

import (
	"context"
	"errors"
	"fmt"
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
	CodeWrongGameStatus
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

type RoundErrorCode int

const (
	CodeLobbyTimeout = iota
	CodeRoundCanceled
)

type RoundError struct {
	Code    RoundErrorCode
	Message string
	Cause   error
}

func (e RoundError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("Error %d: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

func (e RoundError) Unwrap() error {
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
	Video    db.Video
	Response chan error
}

type VoteSubmission struct {
	Vote     db.Vote
	Response chan error
}

type RoundConfig struct {
	MinPlayers       int
	MaxPlayers       int
	LobbyTimeout     time.Duration
	VotingTimeout    time.Duration
	AddPlayerTimeout time.Duration
	AddVideoTimeout  time.Duration
	AddVoteTimeout   time.Duration
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

func (r *Round) Run() {
	defer r.Cancel()

	r.updateGameStatus(db.GameStatusLobby)
	err := r.LobbyStage()
	if err != nil {
		r.completeRound(err)
		return
	}

	r.updateGameStatus(db.GameStatusVoting)
	err = r.VotingStage()
	if err != nil {
		r.completeRound(err)
		return
	}

	r.completeRound(nil)
}

func (r *Round) updateGameStatus(status db.GameStatus) {
	r.StatusUpdate <- GameStatusUpdate{
		GameID: r.Game.ID,
		Status: status,
		Error:  nil,
	}
}

func (r *Round) completeRound(err error) {
	r.StatusUpdate <- GameStatusUpdate{
		GameID: r.Game.ID,
		Status: db.GameStatusComplete,
		Error:  err,
	}

	close(r.StatusUpdate)
}

func (r *Round) LobbyStage() error {
	timer := time.NewTimer(r.Config.LobbyTimeout)
	defer timer.Stop()

	for {
		select {
		case <-r.Ctx.Done():
			return RoundError{
				Code: CodeRoundCanceled,
			}
		case p := <-r.PlayerJoin:
			r.Players = append(r.Players, p.Player)
			p.Response <- nil
			close(p.Response)
		case v := <-r.VideoSubmission:
			r.Videos = append(r.Videos, v.Video)
			v.Response <- nil
			close(v.Response)
		case <-timer.C:
			if len(r.Players) >= r.Config.MinPlayers && len(r.Videos) == len(r.Players) {
				return nil
			}

			return RoundError{Code: CodeLobbyTimeout}
		}
	}
}

func (r *Round) VotingStage() error {
	timer := time.NewTimer(r.Config.LobbyTimeout)
	defer timer.Stop()

	for {
		select {
		case <-r.Ctx.Done():
			return RoundError{
				Code: CodeRoundCanceled,
			}
		case v := <-r.VoteSubmission:
			r.Votes = append(r.Votes, v.Vote)
			v.Response <- nil
			close(v.Response)
		case <-timer.C:
			return nil
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
		MinPlayers:       1,
		MaxPlayers:       8,
		LobbyTimeout:     30 * time.Second,
		VotingTimeout:    30 * time.Second,
		AddPlayerTimeout: 3 * time.Second,
		AddVideoTimeout:  3 * time.Second,
		AddVoteTimeout:   3 * time.Second,
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
	round, err := g.getRound(gameId)
	if err != nil {
		return err
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

		if game.Status != db.GameStatusLobby {
			return nil, GameServiceError{
				Code:    CodeWrongGameStatus,
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

	err = g.submitWithTimeout(
		round.Config.AddPlayerTimeout,
		func(response chan error) {
			round.PlayerJoin <- PlayerJoinRequest{Player: *player, Response: response}
		},
		"can't add player",
	)

	return err
}

func (g *GameService) SubmitVideo(ctx context.Context, params db.CreateVideoParams) (*db.Video, error) {
	round, err := g.getRound(params.GameID)
	if err != nil {
		return nil, err
	}

	video, err := db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*db.Video, error) {
		q := g.txm.Querier(tx)
		err := g.checkPlayerInGameAndGameStatus(ctx, params.GameID, params.PlayerID, db.GameStatusLobby, q)
		if err != nil {
			return nil, err
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

	err = g.submitWithTimeout(
		round.Config.AddVideoTimeout,
		func(response chan error) {
			round.VideoSubmission <- VideoSubmission{Video: *video, Response: response}
		},
		"can't add video",
	)

	if err != nil {
		return nil, err
	}

	return video, nil
}

func (g *GameService) SubmitVote(ctx context.Context, params db.CreateVoteParams) (*db.Vote, error) {
	round, err := g.getRound(params.GameID)
	if err != nil {
		return nil, err
	}

	vote, err := db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*db.Vote, error) {
		q := g.txm.Querier(tx)
		err := g.checkPlayerInGameAndGameStatus(ctx, params.GameID, params.VoterID, db.GameStatusVoting, q)
		if err != nil {
			return nil, err
		}

		vote, err := q.CreateVote(ctx, params)
		if err != nil {
			return nil, GameServiceError{
				Code:    CodeDbError,
				Message: "can't create vote",
				Cause:   err,
			}
		}

		return &vote, nil
	})

	if err != nil {
		return nil, err
	}

	err = g.submitWithTimeout(
		round.Config.AddVoteTimeout,
		func(response chan error) {
			round.VoteSubmission <- VoteSubmission{Vote: *vote, Response: response}
		},
		"can't add vote",
	)

	if err != nil {
		return nil, err
	}

	return vote, nil
}

func (g *GameService) getRound(gameId int64) (*Round, error) {
	round, ok := g.rounds[gameId]
	if !ok {
		return nil, GameServiceError{
			Code:    CodeNotFound,
			Message: "round not found",
		}
	}

	return round, nil
}

func (g *GameService) checkPlayerInGameAndGameStatus(ctx context.Context, gameId int64, playerId int64, status db.GameStatus, q db.Querier) error {
	ok, err := q.IsPlayerInGame(ctx, db.IsPlayerInGameParams{
		GameID:   gameId,
		PlayerID: playerId,
	})

	if err != nil {
		return GameServiceError{
			Code:    getErrorCode(err),
			Message: "can't find player in game",
			Cause:   err,
		}
	}

	if !ok {
		return GameServiceError{
			Code:    CodeNotFound,
			Message: "player not in game",
		}
	}

	game, err := q.GetGame(ctx, db.GetGameParams{ID: gameId})
	if err != nil {
		return GameServiceError{
			Code:    getErrorCode(err),
			Message: "can't fetch game",
			Cause:   err,
		}
	}

	if game.Status != status {
		return GameServiceError{
			Code:    CodeWrongGameStatus,
			Message: "game completed or not started",
		}
	}

	return nil
}

func (g *GameService) submitWithTimeout(timeout time.Duration, sendFunc func(response chan error), errMsg string) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	response := make(chan error, 1)
	sendFunc(response)

	select {
	case err := <-response:
		if err != nil {
			return GameServiceError{
				Code:    CodeUnknown,
				Message: errMsg,
				Cause:   err,
			}
		}
		return nil
	case <-timer.C:
		return GameServiceError{
			Code:    CodeTimeout,
			Message: errMsg,
		}
	}
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

func getErrorCode(err error) GameServiceErrorCode {
	if errors.Is(err, pgx.ErrNoRows) {
		return CodeNotFound
	}

	return CodeDbError
}

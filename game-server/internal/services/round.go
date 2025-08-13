package services

import (
	"context"
	"fmt"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/jackc/pgx/v5"
)

type RoundErrorCode int

const (
	RoundErrorUnknownCode = iota
	RoundErrorLobbyTimeoutCode
	RoundErrorCanceledCode
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

type PlayerJoin struct {
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
	txm             db.TxManager
	Game            db.Game
	Config          RoundConfig
	Players         []db.Player
	Videos          []db.Video
	Votes           []db.Vote
	PlayerJoin      chan PlayerJoin
	VideoSubmission chan VideoSubmission
	VoteSubmission  chan VoteSubmission
	StatusUpdate    chan GameStatusUpdate
	Ctx             context.Context
	Cancel          context.CancelFunc
}

func NewRound(txm db.TxManager, game db.Game, config RoundConfig) *Round {
	ctx, cancel := context.WithCancel(context.Background())
	return &Round{
		txm:             txm,
		Game:            game,
		Config:          config,
		PlayerJoin:      make(chan PlayerJoin),
		StatusUpdate:    make(chan GameStatusUpdate),
		VideoSubmission: make(chan VideoSubmission),
		VoteSubmission:  make(chan VoteSubmission),
		Ctx:             ctx,
		Cancel:          cancel,
	}
}

func (r *Round) Run() {
	defer r.Cancel()

	err := r.updateGameStatus(db.GameStatusLobby, nil)
	if err != nil {
		return
	}

	err = r.lobbyStage()
	if err != nil {
		_ = r.updateGameStatus(db.GameStatusComplete, err)
		return
	}

	_ = r.updateGameStatus(db.GameStatusComplete, nil)
	close(r.StatusUpdate)
}

func (r *Round) updateGameStatus(status db.GameStatus, stErr error) error {
	err := db.WithTx(r.Ctx, r.txm, func(ctx context.Context, tx pgx.Tx) error {
		q := r.txm.Querier(tx)
		return q.UpdateGameStatus(ctx, db.UpdateGameStatusParams{ID: r.Game.ID, Status: status})
	})

	if err != nil {
		r.StatusUpdate <- GameStatusUpdate{
			GameID: r.Game.ID,
			Status: status,
			Error:  RoundError{
				Code: RoundErrorUnknownCode,
				Message: "failed to update game status",
				Cause: err,
			},
		}
		return err
	}

	r.StatusUpdate <- GameStatusUpdate{
		GameID: r.Game.ID,
		Status: status,
		Error:  stErr,
	}

	return nil
}

func (r *Round) lobbyStage() error {
	timer := time.NewTimer(r.Config.LobbyTimeout)
	defer timer.Stop()

	for {
		select {
		case <-r.Ctx.Done():
			return RoundError{
				Code: RoundErrorCanceledCode,
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

			return RoundError{Code: RoundErrorLobbyTimeoutCode}
		}
	}
}

// func (r *Round) VotingStage() error {
// 	timer := time.NewTimer(r.Config.LobbyTimeout)
// 	defer timer.Stop()

// 	for {
// 		select {
// 		case <-r.Ctx.Done():
// 			return RoundError{
// 				Code: RoundErrorCanceledCode,
// 			}
// 		case v := <-r.VoteSubmission:
// 			r.Votes = append(r.Votes, v.Vote)
// 			v.Response <- nil
// 			close(v.Response)
// 		case <-timer.C:
// 			return nil
// 		}
// 	}
// }

package services

import (
	"context"
	"fmt"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/utils"
	"github.com/jackc/pgx/v5"
)

type RoundErrorCode int

const (
	RoundErrUnknown = iota
	RoundErrCanceled
	RoundErrDbError
	RoundErrNotFound
	RoundErrConstraintViolation
	RoundErrTooManyPlayers
	RoundErrWrongGameState
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

type Countdown struct {
	timer    *time.Timer
	started  time.Time
	duration time.Duration
}

func NewCountdown(d time.Duration) Countdown {
	return Countdown{
		timer:    time.NewTimer(d),
		started:  time.Now(),
		duration: d,
	}
}

func (c *Countdown) Remaining() time.Duration {
	if c.timer == nil {
		return 0
	}

	r := c.duration - time.Since(c.started)
	if r < 0 {
		return 0
	}

	return r
}

func (c *Countdown) Reset(d time.Duration) {
	if !c.timer.Stop() {
		select {
		case <-c.timer.C:
		default:
		}
	}
	c.timer.Reset(d)
	c.started = time.Now()
	c.duration = d
}

func (c *Countdown) Stop() {
	if c.timer != nil {
		c.timer.Stop()
	}
}

type GameStatusUpdate struct {
	GameID int64
	Status db.GameStatus
	Error  error
}

type PlayerJoin struct {
	Player   db.AddPlayerToGameParams
	Response chan error
}

type VideoSubmission struct {
	Video    db.CreateVideoParams
	Response chan error
}

type VoteSubmission struct {
	Vote     db.Vote
	Response chan error
}

type RoundConfig struct {
	MinPlayers        int
	MaxPlayers        int
	LobbyTimeout      time.Duration
	SubmittingTimeout time.Duration
	VotingTimeout     time.Duration
	AddPlayerTimeout  time.Duration
	AddVideoTimeout   time.Duration
	AddVoteTimeout    time.Duration
}

type Round struct {
	txm             db.TxManager
	Game            db.Game
	Config          RoundConfig
	StageCountdown  Countdown
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
	r.gameLoop()
}

func (r *Round) updateGameStatus(status db.GameStatus) error {
	game, err := db.WithTxValue(r.Ctx, r.txm, func(ctx context.Context, tx pgx.Tx) (db.Game, error) {
		q := r.txm.Querier(tx)
		return q.UpdateGameStatus(ctx, db.UpdateGameStatusParams{ID: r.Game.ID, Status: status})
	})

	if err != nil {
		return RoundError{
			Code:    RoundErrUnknown,
			Message: "failed to update game status",
			Cause:   err,
		}
	}

	r.Game = game

	return nil
}

func (r *Round) updateRound(err error) {
	r.StatusUpdate <- GameStatusUpdate{
		GameID: r.Game.ID,
		Status: r.Game.Status,
		Error:  err,
	}
}

func (r *Round) gameLoop() {
	defer r.StageCountdown.Stop()

	err := r.toLobbyStage()
	if err != nil {
		return
	}

Loop:
	for {
		select {
		case <-r.Ctx.Done():
			break Loop
		case p := <-r.PlayerJoin:
			_ = r.addPlayer(p)
			if r.checkEnoughtPlayers() {
				r.toSubmittingStage()
			}
		case v := <-r.VideoSubmission:
			r.addVideo(v)
		case v := <-r.VoteSubmission:
			r.addVote(v)
		case <-r.StageCountdown.timer.C:
			err := r.nextStage()
			if err != nil {
				break Loop
			}
		}

		if r.Game.Status == db.GameStatusWinner {
			break Loop
		}
	}

	r.toCompleteStage()
}

func (r *Round) nextStage() error {
	switch r.Game.Status {
	case db.GameStatusLobby:
		return r.toSubmittingStage()
	case db.GameStatusSubmitting:
		return r.toVotingStage()
	case db.GameStatusVoting:
		return r.toWinnerStage()
	}

	err := r.updateGameStatus(db.GameStatusWinner)
	r.updateRound(err)

	return err
}

func (r *Round) toLobbyStage() error {
	err := r.updateGameStatus(db.GameStatusLobby)
	r.updateRound(err)
	if err != nil {
		return err
	}
	r.StageCountdown.Stop()
	r.StageCountdown = NewCountdown(r.Config.LobbyTimeout)

	return nil
}

func (r *Round) toSubmittingStage() error {
	err := r.updateGameStatus(db.GameStatusSubmitting)
	r.updateRound(err)
	if err != nil {
		return err
	}
	r.StageCountdown.Stop()
	r.StageCountdown = NewCountdown(r.Config.SubmittingTimeout)

	return nil
}

func (r *Round) toVotingStage() error {
	err := r.updateGameStatus(db.GameStatusVoting)
	r.updateRound(err)
	if err != nil {
		return err
	}
	r.StageCountdown.Stop()
	r.StageCountdown = NewCountdown(r.Config.VotingTimeout)

	return nil
}

func (r *Round) toWinnerStage() error {
	err := r.updateGameStatus(db.GameStatusWinner)
	r.updateRound(err)
	return err
}

func (r *Round) toCompleteStage() {
	err := r.updateGameStatus(db.GameStatusComplete)
	r.updateRound(err)
	close(r.StatusUpdate)
}

func (r *Round) addPlayer(pj PlayerJoin) error {
	if r.Game.Status != db.GameStatusLobby {
		err := RoundError{
			Code: RoundErrWrongGameState,
		}

		responseAndClose(pj.Response, err)
		return err
	}

	for _, p := range r.Players {
		if p.ID == pj.Player.PlayerID {
			responseAndClose(pj.Response, nil)
			return nil
		}
	}

	if len(r.Players) > r.Config.MaxPlayers {
		err := RoundError{
			Code: RoundErrTooManyPlayers,
		}

		responseAndClose(pj.Response, err)
		return err
	}

	player, err := db.WithTxValue(context.Background(), r.txm, func(ctx context.Context, tx pgx.Tx) (*db.Player, error) {
		q := r.txm.Querier(tx)

		err := q.AddPlayerToGame(ctx, pj.Player)

		if err != nil {
			if utils.IsClass23(err) {
				return nil, RoundError{
					Code:    RoundErrConstraintViolation,
					Message: "constrain violation adding player",
					Cause:   err,
				}
			}

			return nil, RoundError{
				Code:    RoundErrDbError,
				Message: "db error adding player",
				Cause:   err,
			}
		}

		player, err := q.GetPlayer(ctx, db.GetPlayerParams{ID: pj.Player.PlayerID})
		if err != nil {
			return nil, RoundError{
				Code:    RoundErrDbError,
				Message: "db error getting player",
				Cause:   err,
			}
		}

		return &player, nil
	})

	if player != nil {
		r.Players = append(r.Players, *player)
	}

	responseAndClose(pj.Response, err)
	return err
}

func (r *Round) checkEnoughtPlayers() bool {
	return r.Game.Status == db.GameStatusLobby && len(r.Players) >= r.Config.MaxPlayers
}

func (r *Round) addVideo(vs VideoSubmission) error {
	if r.Game.Status != db.GameStatusLobby && r.Game.Status != db.GameStatusSubmitting {
		err := RoundError{
			Code: RoundErrWrongGameState,
		}

		responseAndClose(vs.Response, err)
		return err
	}

	videos, err := db.WithTxValue(context.Background(), r.txm, func(ctx context.Context, tx pgx.Tx) ([]db.Video, error) {
		q := r.txm.Querier(tx)

		video, err := q.CreateVideo(ctx, vs.Video)
		if err != nil {
			if utils.IsClass23(err) {
				return nil, RoundError{
					Code:    RoundErrConstraintViolation,
					Message: "constrain violation adding video",
					Cause:   err,
				}
			}

			return nil, RoundError{
				Code:    RoundErrDbError,
				Message: "can't create video",
				Cause:   err,
			}
		}

		err = q.InvalidateOtherVideos(ctx, db.InvalidateOtherVideosParams{
			GameID: vs.Video.GameID, 
			PlayerID: vs.Video.PlayerID, 
			ID: video.ID,
		})

		if err != nil {
			return nil, RoundError{
				Code:    RoundErrDbError,
				Message: "can't create video",
				Cause:   err,
			}
		}

		videos, err := q.GetVideosByGame(ctx, db.GetVideosByGameParams{GameID: r.Game.ID})
		if err != nil {
			return nil, RoundError{
				Code:    RoundErrDbError,
				Message: "can't create video",
				Cause:   err,
			}
		}

		return videos, nil
	})

	if videos != nil {
		r.Videos = videos
	}

	responseAndClose(vs.Response, err)
	return nil
}

func (r *Round) addVote(v VoteSubmission) error {
	r.Votes = append(r.Votes, v.Vote)
	v.Response <- nil
	close(v.Response)
	return nil
}

func responseAndClose(re chan error, err error) {
	re <- err
	close(re)
}

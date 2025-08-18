package services

import (
	"context"
	"fmt"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/antonaby/shortsbattle/game-server/internal/utils"
	"github.com/jackc/pgx/v5"
)

type GameInstanceErrorCode int

const (
	GIErrUnknown = iota
	GIErrCanceled
	GIErrDbError
	GIErrNotFound
	GIErrConstraintViolation
	GIErrTooManyPlayers
	GIErrWrongGameState
	GIErrPublicationFailed
)

type GameInstanceError struct {
	Code    GameInstanceErrorCode
	Message string
	Cause   error
}

func (e GameInstanceError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("Error %d: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

func (e GameInstanceError) Unwrap() error {
	return e.Cause
}

type GameEventPublisher interface {
	PublishGameUpdate(channel string, upd models.GameUpdate) error
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

type PlayerJoin struct {
	Player   db.AddPlayerToGameParams
	Ctx      context.Context
	Response chan error
}

type VideoSubmission struct {
	Video    db.CreateVideoParams
	Response chan error
}

type VoteSubmission struct {
	Vote     db.CreateVoteParams
	Response chan error
}

type GameInstanceConfig struct {
	MinPlayers        int
	MaxPlayers        int
	LobbyTimeout      time.Duration
	SubmittingTimeout time.Duration
	VotingTimeout     time.Duration
	AddPlayerTimeout  time.Duration
	AddVideoTimeout   time.Duration
	AddVoteTimeout    time.Duration
}

type GameInstance struct {
	txm             db.TxManager
	publisher       GameEventPublisher
	Game            db.Game
	Config          GameInstanceConfig
	StageCountdown  Countdown
	Players         []db.Player
	Videos          []db.Video
	Votes           []db.Vote
	PlayerJoin      chan PlayerJoin
	VideoSubmission chan VideoSubmission
	VoteSubmission  chan VoteSubmission
	Ctx             context.Context
	Cancel          context.CancelFunc
}

func NewGameInstance(txm db.TxManager, publisher GameEventPublisher, game db.Game, config GameInstanceConfig) *GameInstance {
	ctx, cancel := context.WithCancel(context.Background())
	return &GameInstance{
		txm:             txm,
		publisher:       publisher,
		Game:            game,
		Config:          config,
		PlayerJoin:      make(chan PlayerJoin),
		VideoSubmission: make(chan VideoSubmission),
		VoteSubmission:  make(chan VoteSubmission),
		Ctx:             ctx,
		Cancel:          cancel,
	}
}

func (r *GameInstance) Run() error {
	defer r.Cancel()
	r.gameLoop()
	return nil
}

func (r *GameInstance) updateGameStatus(status db.GameStatus) error {
	game, err := db.WithTxValue(r.Ctx, r.txm, func(ctx context.Context, tx pgx.Tx) (db.Game, error) {
		q := r.txm.Querier(tx)
		return q.UpdateGameStatus(ctx, db.UpdateGameStatusParams{ID: r.Game.ID, Status: status})
	})

	if err != nil {
		return GameInstanceError{
			Code:    GIErrUnknown,
			Message: "failed to update game status",
			Cause:   err,
		}
	}

	r.Game = game

	statusUpdate := models.GameUpdate{
		ID: game.ID,
		Status: game.Status,
	}

	err = r.publisher.PublishGameUpdate(fmt.Sprintf("game_%d", r.Game.ID), statusUpdate)
	if err != nil {
		return GameInstanceError{
			Code:    GIErrPublicationFailed,
			Message: "failed to publish game status",
			Cause:   err,
		}
	}

	return nil
}

func (r *GameInstance) gameLoop() {
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

func (r *GameInstance) nextStage() error {
	switch r.Game.Status {
	case db.GameStatusLobby:
		return r.toSubmittingStage()
	case db.GameStatusSubmitting:
		return r.toVotingStage()
	case db.GameStatusVoting:
		return r.toWinnerStage()
	}

	err := r.updateGameStatus(db.GameStatusWinner)

	return err
}

func (r *GameInstance) toLobbyStage() error {
	err := r.updateGameStatus(db.GameStatusLobby)
	if err != nil {
		return err
	}
	r.StageCountdown.Stop()
	r.StageCountdown = NewCountdown(r.Config.LobbyTimeout)

	return nil
}

func (r *GameInstance) toSubmittingStage() error {
	err := r.updateGameStatus(db.GameStatusSubmitting)
	if err != nil {
		return err
	}
	r.StageCountdown.Stop()
	r.StageCountdown = NewCountdown(r.Config.SubmittingTimeout)

	return nil
}

func (r *GameInstance) toVotingStage() error {
	err := r.updateGameStatus(db.GameStatusVoting)
	if err != nil {
		return err
	}
	r.StageCountdown.Stop()
	r.StageCountdown = NewCountdown(r.Config.VotingTimeout)

	return nil
}

func (r *GameInstance) toWinnerStage() error {
	err := r.updateGameStatus(db.GameStatusWinner)
	return err
}

func (r *GameInstance) toCompleteStage() {
	_ = r.updateGameStatus(db.GameStatusComplete)
}

func (r *GameInstance) addPlayer(pj PlayerJoin) error {
	if r.Game.Status != db.GameStatusCreated && r.Game.Status != db.GameStatusLobby {
		err := GameInstanceError{
			Code: GIErrWrongGameState,
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
		err := GameInstanceError{
			Code: GIErrTooManyPlayers,
		}

		responseAndClose(pj.Response, err)
		return err
	}

	player, err := db.WithTxValue(pj.Ctx, r.txm, func(ctx context.Context, tx pgx.Tx) (*db.Player, error) {
		q := r.txm.Querier(tx)

		err := q.AddPlayerToGame(ctx, pj.Player)

		if err != nil {
			if utils.IsClass23(err) {
				return nil, GameInstanceError{
					Code:    GIErrConstraintViolation,
					Message: "constrain violation adding player",
					Cause:   err,
				}
			}

			return nil, GameInstanceError{
				Code:    GIErrDbError,
				Message: "db error adding player",
				Cause:   err,
			}
		}

		player, err := q.GetPlayer(ctx, db.GetPlayerParams{ID: pj.Player.PlayerID})
		if err != nil {
			return nil, GameInstanceError{
				Code:    GIErrDbError,
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

func (r *GameInstance) checkEnoughtPlayers() bool {
	return r.Game.Status == db.GameStatusLobby && len(r.Players) >= r.Config.MaxPlayers
}

func (r *GameInstance) addVideo(vs VideoSubmission) error {
	if r.Game.Status != db.GameStatusLobby && r.Game.Status != db.GameStatusSubmitting {
		err := GameInstanceError{
			Code: GIErrWrongGameState,
		}

		responseAndClose(vs.Response, err)
		return err
	}

	videos, err := db.WithTxValue(context.Background(), r.txm, func(ctx context.Context, tx pgx.Tx) ([]db.Video, error) {
		q := r.txm.Querier(tx)

		vs.Video.IsActual = true
		video, err := q.CreateVideo(ctx, vs.Video)
		if err != nil {
			if utils.IsClass23(err) {
				return nil, GameInstanceError{
					Code:    GIErrConstraintViolation,
					Message: "constrain violation adding video",
					Cause:   err,
				}
			}

			return nil, GameInstanceError{
				Code:    GIErrDbError,
				Message: "can't create video",
				Cause:   err,
			}
		}

		err = q.InvalidateOtherVideos(ctx, db.InvalidateOtherVideosParams{
			GameID:   vs.Video.GameID,
			PlayerID: vs.Video.PlayerID,
			ID:       video.ID,
		})

		if err != nil {
			return nil, GameInstanceError{
				Code:    GIErrDbError,
				Message: "can't create video",
				Cause:   err,
			}
		}

		videos, err := q.GetVideosByGame(ctx, db.GetVideosByGameParams{GameID: r.Game.ID})
		if err != nil {
			return nil, GameInstanceError{
				Code:    GIErrDbError,
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

func (r *GameInstance) addVote(vs VoteSubmission) error {
	if r.Game.Status != db.GameStatusVoting {
		err := GameInstanceError{
			Code: GIErrWrongGameState,
		}

		responseAndClose(vs.Response, err)
		return err
	}

	votes, err := db.WithTxValue(context.Background(), r.txm, func(ctx context.Context, tx pgx.Tx) ([]db.Vote, error) {
		q := r.txm.Querier(tx)

		vs.Vote.IsActual = true
		vote, err := q.CreateVote(ctx, vs.Vote)
		if err != nil {
			if utils.IsClass23(err) {
				return nil, GameInstanceError{
					Code:    GIErrConstraintViolation,
					Message: "constrain violation adding vote",
					Cause:   err,
				}
			}

			return nil, GameInstanceError{
				Code:    GIErrDbError,
				Message: "can't create vote",
				Cause:   err,
			}
		}

		err = q.InvalidateOtherVotes(ctx, db.InvalidateOtherVotesParams{
			GameID:  vs.Vote.GameID,
			VoterID: vs.Vote.VoterID,
			VideoID: vs.Vote.VideoID,
			ID:      vote.ID,
		})

		if err != nil {
			return nil, GameInstanceError{
				Code:    GIErrDbError,
				Message: "can't create vote",
				Cause:   err,
			}
		}

		votes, err := q.GetVotesByGame(ctx, db.GetVotesByGameParams{GameID: r.Game.ID})
		if err != nil {
			return nil, GameInstanceError{
				Code:    GIErrDbError,
				Message: "can't create vote",
				Cause:   err,
			}
		}

		return votes, nil
	})

	if votes != nil {
		r.Votes = votes
	}

	responseAndClose(vs.Response, err)
	return nil
}

func responseAndClose(re chan error, err error) {
	re <- err
	close(re)
}

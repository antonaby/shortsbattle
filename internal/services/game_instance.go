package services

import (
	"context"
)

type Game struct {
	Config GameConfig
	Ctx    context.Context
	Cancel context.CancelFunc
}

// func NewGameInstance(txm db.TxManager, publisher GameEventPublisher, game db.Game, config GameInstanceConfig) *GameInstance {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	return &GameInstance{
// 		txm:             txm,
// 		publisher:       publisher,
// 		Game:            game,
// 		Config:          config,
// 		PlayerJoin:      make(chan PlayerJoin),
// 		VideoSubmission: make(chan VideoSubmission),
// 		VoteSubmission:  make(chan VoteSubmission),
// 		Ctx:             ctx,
// 		Cancel:          cancel,
// 	}
// }

// func (r *GameInstance) Run() error {
// 	defer r.Cancel()
// 	r.gameLoop()
// 	return nil
// }

// func (r *GameInstance) updateGameStatus(status db.GameStatus) error {
// 	game, err := db.WithTxValue(r.Ctx, r.txm, func(ctx context.Context, tx pgx.Tx) (db.Game, error) {
// 		q := r.txm.Querier(tx)
// 		return q.UpdateGameStatus(ctx, db.UpdateGameStatusParams{ID: r.Game.ID, Status: status})
// 	})

// 	if err != nil {
// 		return GameInstanceError{
// 			Code:    GIErrUnknown,
// 			Message: "failed to update game status",
// 			Cause:   err,
// 		}
// 	}

// 	r.Game = game

// 	return r.publishGameStatus()
// }

// func (r *GameInstance) publishGameStatus() error {
// 	statusUpdate := models.GameUpdate{
// 		ID:                 r.Game.ID,
// 		Status:             r.Game.Status,
// 		StageTimeRemaining: r.StageCountdown.Remaining().Milliseconds(),
// 	}

// 	err := r.publisher.PublishGameUpdate(fmt.Sprintf("game_%d", r.Game.ID), statusUpdate)
// 	if err != nil {
// 		return GameInstanceError{
// 			Code:    GIErrPublicationFailed,
// 			Message: "failed to publish game status",
// 			Cause:   err,
// 		}
// 	}

// 	return nil
// }

// func (r *GameInstance) gameLoop() {
// 	ticker := time.NewTicker(r.Config.TickerDuration)
// 	defer ticker.Stop()

// 	defer r.StageCountdown.Stop()

// 	err := r.setCreatedTimer()
// 	if err != nil {
// 		return
// 	}

// Loop:
// 	for {
// 		select {
// 		case <-r.Ctx.Done():
// 			break Loop
// 		case <-ticker.C:
// 			_ = r.publishGameStatus()
// 		case p := <-r.PlayerJoin:
// 			_ = r.addPlayer(p)
// 			if r.checkEnoughtPlayers() {
// 				r.toSubmittingStage()
// 			}
// 		case v := <-r.VideoSubmission:
// 			r.addVideo(v)
// 		case v := <-r.VoteSubmission:
// 			r.addVote(v)
// 		case <-r.StageCountdown.timer.C:
// 			err := r.nextStage()
// 			if err != nil {
// 				break Loop
// 			}
// 		}

// 		if r.Game.Status == db.GameStatusWinner {
// 			break Loop
// 		}
// 	}

// 	r.toCompleteStage()
// }

// func (r *GameInstance) nextStage() error {
// 	switch r.Game.Status {
// 	case db.GameStatusCreated:
// 		return r.toLobbyStage()
// 	case db.GameStatusLobby:
// 		return r.toSubmittingStage()
// 	case db.GameStatusSubmitting:
// 		return r.toVotingStage()
// 	case db.GameStatusVoting:
// 		return r.toWinnerStage()
// 	}

// 	err := r.updateGameStatus(db.GameStatusWinner)

// 	return err
// }

// func (r *GameInstance) setCreatedTimer() error {
// 	r.StageCountdown = NewCountdown(r.Config.CreatedTimeout)

// 	err := r.updateGameStatus(db.GameStatusCreated)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (r *GameInstance) toLobbyStage() error {
// 	r.StageCountdown.Reset(r.Config.LobbyTimeout)

// 	err := r.updateGameStatus(db.GameStatusLobby)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (r *GameInstance) toSubmittingStage() error {
// 	r.StageCountdown.Reset(r.Config.SubmittingTimeout)

// 	err := r.updateGameStatus(db.GameStatusSubmitting)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (r *GameInstance) toVotingStage() error {
// 	r.StageCountdown.Reset(r.Config.VotingTimeout)

// 	err := r.updateGameStatus(db.GameStatusVoting)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (r *GameInstance) toWinnerStage() error {
// 	r.StageCountdown.Stop()
// 	err := r.updateGameStatus(db.GameStatusWinner)
// 	return err
// }

// func (r *GameInstance) toCompleteStage() {
// 	r.StageCountdown.Stop()
// 	_ = r.updateGameStatus(db.GameStatusComplete)
// }

// // TODO: fix multiple joining players
// func (r *GameInstance) addPlayer(pj PlayerJoin) error {
// 	if r.Game.Status != db.GameStatusCreated && r.Game.Status != db.GameStatusLobby {
// 		err := GameInstanceError{
// 			Code: GIErrWrongGameState,
// 		}

// 		responseAndClose(pj.Response, err)
// 		return err
// 	}

// 	for _, p := range r.Players {
// 		if p.ID == pj.Player.PlayerID {
// 			responseAndClose(pj.Response, nil)
// 			return nil
// 		}
// 	}

// 	if len(r.Players) > r.Config.MaxPlayers {
// 		err := GameInstanceError{
// 			Code: GIErrTooManyPlayers,
// 		}

// 		responseAndClose(pj.Response, err)
// 		return err
// 	}

// 	player, err := db.WithTxValue(pj.Ctx, r.txm, func(ctx context.Context, tx pgx.Tx) (*db.Player, error) {
// 		q := r.txm.Querier(tx)

// 		err := q.AddPlayerToGame(ctx, pj.Player)

// 		if err != nil {
// 			if utils.IsClass23(err) {
// 				return nil, GameInstanceError{
// 					Code:    GIErrConstraintViolation,
// 					Message: "constrain violation adding player",
// 					Cause:   err,
// 				}
// 			}

// 			return nil, GameInstanceError{
// 				Code:    GIErrDbError,
// 				Message: "db error adding player",
// 				Cause:   err,
// 			}
// 		}

// 		player, err := q.GetPlayer(ctx, db.GetPlayerParams{ID: pj.Player.PlayerID})
// 		if err != nil {
// 			return nil, GameInstanceError{
// 				Code:    GIErrDbError,
// 				Message: "db error getting player",
// 				Cause:   err,
// 			}
// 		}

// 		return &player, nil
// 	})

// 	if player != nil {
// 		r.Players = append(r.Players, *player)
// 	}

// 	responseAndClose(pj.Response, err)
// 	return err
// }

// func (r *GameInstance) checkEnoughtPlayers() bool {
// 	return r.Game.Status == db.GameStatusLobby && len(r.Players) >= r.Config.MaxPlayers
// }

// func (r *GameInstance) addVideo(vs VideoSubmission) error {
// 	if r.Game.Status != db.GameStatusLobby && r.Game.Status != db.GameStatusSubmitting {
// 		err := GameInstanceError{
// 			Code: GIErrWrongGameState,
// 		}

// 		responseAndClose(vs.Response, err)
// 		return err
// 	}

// 	videos, err := db.WithTxValue(vs.Ctx, r.txm, func(ctx context.Context, tx pgx.Tx) ([]db.Video, error) {
// 		q := r.txm.Querier(tx)

// 		vs.Video.IsActual = true
// 		video, err := q.CreateVideo(ctx, vs.Video)
// 		if err != nil {
// 			if utils.IsClass23(err) {
// 				return nil, GameInstanceError{
// 					Code:    GIErrConstraintViolation,
// 					Message: "constrain violation adding video",
// 					Cause:   err,
// 				}
// 			}

// 			return nil, GameInstanceError{
// 				Code:    GIErrDbError,
// 				Message: "can't create video",
// 				Cause:   err,
// 			}
// 		}

// 		err = q.InvalidateOtherVideos(ctx, db.InvalidateOtherVideosParams{
// 			GameID:   vs.Video.GameID,
// 			PlayerID: vs.Video.PlayerID,
// 			ID:       video.ID,
// 		})

// 		if err != nil {
// 			return nil, GameInstanceError{
// 				Code:    GIErrDbError,
// 				Message: "can't create video",
// 				Cause:   err,
// 			}
// 		}

// 		videos, err := q.GetVideosByGame(ctx, db.GetVideosByGameParams{GameID: r.Game.ID})
// 		if err != nil {
// 			return nil, GameInstanceError{
// 				Code:    GIErrDbError,
// 				Message: "can't create video",
// 				Cause:   err,
// 			}
// 		}

// 		return videos, nil
// 	})

// 	if videos != nil {
// 		r.Videos = videos
// 	}

// 	responseAndClose(vs.Response, err)
// 	return nil
// }

// func (r *GameInstance) addVote(vs VoteSubmission) error {
// 	if r.Game.Status != db.GameStatusVoting {
// 		err := GameInstanceError{
// 			Code: GIErrWrongGameState,
// 		}

// 		responseAndClose(vs.Response, err)
// 		return err
// 	}

// 	votes, err := db.WithTxValue(context.Background(), r.txm, func(ctx context.Context, tx pgx.Tx) ([]db.Vote, error) {
// 		q := r.txm.Querier(tx)

// 		vs.Vote.IsActual = true
// 		vote, err := q.CreateVote(ctx, vs.Vote)
// 		if err != nil {
// 			if utils.IsClass23(err) {
// 				return nil, GameInstanceError{
// 					Code:    GIErrConstraintViolation,
// 					Message: "constrain violation adding vote",
// 					Cause:   err,
// 				}
// 			}

// 			return nil, GameInstanceError{
// 				Code:    GIErrDbError,
// 				Message: "can't create vote",
// 				Cause:   err,
// 			}
// 		}

// 		err = q.InvalidateOtherVotes(ctx, db.InvalidateOtherVotesParams{
// 			GameID:  vs.Vote.GameID,
// 			VoterID: vs.Vote.VoterID,
// 			VideoID: vs.Vote.VideoID,
// 			ID:      vote.ID,
// 		})

// 		if err != nil {
// 			return nil, GameInstanceError{
// 				Code:    GIErrDbError,
// 				Message: "can't create vote",
// 				Cause:   err,
// 			}
// 		}

// 		votes, err := q.GetVotesByGame(ctx, db.GetVotesByGameParams{GameID: r.Game.ID})
// 		if err != nil {
// 			return nil, GameInstanceError{
// 				Code:    GIErrDbError,
// 				Message: "can't create vote",
// 				Cause:   err,
// 			}
// 		}

// 		return votes, nil
// 	})

// 	if votes != nil {
// 		r.Votes = votes
// 	}

// 	responseAndClose(vs.Response, err)
// 	return nil
// }

// func responseAndClose(re chan error, err error) {
// 	re <- err
// 	close(re)
// }

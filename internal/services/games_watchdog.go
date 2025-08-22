package services

import (
	"context"

	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
)

type GamesWatchdog struct {
	txm      db.TxManager
	interval time.Duration
	config   GameConfig
}

func NewGamesWatchdog(txm db.TxManager, interval time.Duration, config GameConfig) *GamesWatchdog {
	return &GamesWatchdog{
		txm:      txm,
		interval: interval,
		config:   config,
	}
}

func (wd *GamesWatchdog) Run(ctx context.Context) error {
	t := time.NewTicker(wd.interval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			wd.checkPendingGames()
		}
	}
}

func (wd *GamesWatchdog) checkPendingGames() {
	err := db.WithTx(context.Background(), wd.txm, func(ctx context.Context, tx pgx.Tx) error {
		q := wd.txm.Querier(tx)
		games, err := q.AdvanceGames(ctx, qg.AdvanceGamesParams{
			PLobbyToSubmitting: pgtype.Interval{
				Microseconds: int64(wd.config.LobbyState.Microseconds()),
				Days:         0,
				Months:       0,
				Valid:        true,
			},
			PSubmittingToWathching: pgtype.Interval{
				Microseconds: int64(wd.config.LobbyState.Microseconds()),
				Days:         0,
				Months:       0,
				Valid:        true,
			},
			PWathchingToCompleted: pgtype.Interval{
				Microseconds: int64(wd.config.LobbyState.Microseconds()),
				Days:         0,
				Months:       0,
				Valid:        true,
			},
			PBatchLimit: 20,
		})

		if len(games) > 0 {
			log.Info().Msgf("Updated games: %d", len(games))
		}

		return err
	})

	if err != nil {
		log.Error().Err(err).Msg("failed to advance game states")
	}
}

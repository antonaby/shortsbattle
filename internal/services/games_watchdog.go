package services

import (
	"context"

	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type GamesWatchdog struct {
	txm      db.TxManager
	rc       *redis.Client
	interval time.Duration
	limit    int32
}

// TODO: close lobby 10 seconds before the actual change
// TODO: send everything to Redis Streams instead of changeing in the stored fucntion (add enqueued and processed columns)
func NewGamesWatchdog(txm db.TxManager, rc *redis.Client, interval time.Duration, limit int32) *GamesWatchdog {
	return &GamesWatchdog{
		txm:      txm,
		rc:       rc,
		interval: interval,
		limit:    limit,
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
		games, err := q.AdvanceGames(ctx, qg.AdvanceGamesParams{Limit: wd.limit})

		if len(games) > 0 {
			log.Info().Msgf("Updated games: %d", len(games))
		}

		return err
	})

	if err != nil {
		log.Error().Err(err).Msg("failed to advance game states")
	}
}

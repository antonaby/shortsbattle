package services

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type GameWatchdog struct {
	txm           db.TxManager
	rc            *redis.Client
	interval      time.Duration
	limit         int32
	streamName    string
	streamMaxLean int64
}

// TODO: send everything to Redis Streams instead of changeing in the stored fucntion (add enqueued and processed columns)
func NewGameWatchdog(
	txm db.TxManager, rc *redis.Client,
	interval time.Duration, limit int32,
	streamName string, streamMaxLean int64) *GameWatchdog {
	return &GameWatchdog{
		txm:           txm,
		rc:            rc,
		interval:      interval,
		limit:         limit,
		streamName:    streamName,
		streamMaxLean: streamMaxLean,
	}
}

func (wd *GameWatchdog) Run(ctx context.Context) error {
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

func (wd *GameWatchdog) checkPendingGames() {
	err := db.WithTx(context.Background(), wd.txm, func(ctx context.Context, tx pgx.Tx) error {
		q := wd.txm.Querier(tx)
		games, err := q.AdvanceGames(ctx, qg.AdvanceGamesParams{Limit: wd.limit})

		if err != nil {
			return err
		}

		if len(games) > 0 {
			_, err = wd.rc.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				for _, game := range games {
					pipe.XAdd(ctx, &redis.XAddArgs{
						Stream: wd.streamName,
						MaxLen: wd.streamMaxLean,
						Approx: true,
						Values: map[string]any{
							"game_id": game.ID,
						},
					})
				}

				return nil
			})

			log.Info().Msgf("Updated games: %d", len(games))
		}

		return err
	})

	if err != nil {
		log.Error().Err(err).Msg("failed to advance game states")
	}
}

type GameStateListener struct {
	rc                *redis.Client
	streamName        string
	consumerGroupName string
}

func NewGameStateListener(rc *redis.Client, streamName, consumerGroupName string) *GameStateListener {
	return &GameStateListener{
		rc:                rc,
		streamName:        streamName,
		consumerGroupName: consumerGroupName,
	}
}

func (gsl *GameStateListener) Listen(ctx context.Context) {
	consumerName := gsl.newConsumerName("worker", "1")

	for {
		select {
		case <-ctx.Done():
			goto CLEANUP
		default:
			res, err := gsl.rc.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    gsl.consumerGroupName,
				Consumer: consumerName,
				Streams:  []string{gsl.streamName, ">"},
				Count:    10,
				Block:    5 * time.Second,
				NoAck:    false,
			}).Result()

			if err == redis.Nil {
				continue
			}

			if err != nil {
				log.Error().Err(err).Msgf("readgroup error: %v")
				continue
			}

			for _, str := range res {
				for _, msg := range str.Messages {
					if err := gsl.handleMessage(msg); err != nil {
						log.Error().Err(err).Msgf("handler failed for %s", msg.ID)
						continue
					}
					if err := gsl.rc.XAck(ctx, gsl.streamName, gsl.consumerGroupName, msg.ID).Err(); err != nil {
						log.Error().Err(err).Msgf("ack failed %s", msg.ID)
					}
				}
			}
		}
	}

CLEANUP:
	delCtx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	err := gsl.rc.XGroupDelConsumer(delCtx, gsl.streamName, gsl.consumerGroupName, consumerName).Err()
	if err != nil {
		log.Error().Err(err).Msgf("failed to delete consumer %s", consumerName)
	}
}

func (gsl *GameStateListener) handleMessage(msg redis.XMessage) error {

	return nil
}

func (gsl *GameStateListener) newConsumerName(prefix, suffix string) string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "unknownhost"
	}
	host = strings.ToLower(host)

	re := regexp.MustCompile(`[^a-z0-9]+`)
	host = strings.Trim(re.ReplaceAllString(host, "-"), "-")

	pid := os.Getpid()
	if prefix == "" {
		prefix = "worker"
	}
	if suffix == "" {
		suffix = "1"
	}
	return fmt.Sprintf("%s-%s-%d-%s", prefix, host, pid, suffix)
}

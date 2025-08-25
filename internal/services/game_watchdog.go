package services

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
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

// TODO: add better error handler
type GameStateListener struct {
	rc                *redis.Client
	gm                *GameManager
	streamName        string
	consumerGroupName string
	count             int64
	block             time.Duration
	consumerName      string
}

func NewGameStateListener(
	rc *redis.Client, gm *GameManager,
	streamName, consumerGroupName string,
	count int64, block time.Duration) *GameStateListener {
	return &GameStateListener{
		rc:                rc,
		gm:                gm,
		streamName:        streamName,
		consumerGroupName: consumerGroupName,
		count:             count,
		block:             block,
		consumerName:      newConsumerName("worker", "1"),
	}
}

func (gsl *GameStateListener) Listen(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			goto CLEANUP
		default:
			res, err := gsl.rc.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    gsl.consumerGroupName,
				Consumer: gsl.consumerName,
				Streams:  []string{gsl.streamName, ">"},
				Count:    gsl.count,
				Block:    gsl.block,
				NoAck:    false,
			}).Result()

			if err == redis.Nil {
				continue
			}

			if err != nil {
				log.Error().Err(err).Msgf("readgroup error %s", gsl.streamName)
				continue
			}

			for _, str := range res {
				for _, msg := range str.Messages {
					if err := gsl.handleMessage(ctx, msg); err != nil {
						log.Error().Err(err).Msgf("handler failed for %s:%s", gsl.streamName, msg.ID)
					}
					if err := gsl.rc.XAck(ctx, gsl.streamName, gsl.consumerGroupName, msg.ID).Err(); err != nil {
						log.Error().Err(err).Msgf("ack failed %s:%s", gsl.streamName, msg.ID)
					}
				}
			}
		}
	}

CLEANUP:
	delCtx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	err := gsl.rc.XGroupDelConsumer(delCtx, gsl.streamName, gsl.consumerGroupName, gsl.consumerName).Err()
	if err != nil {
		log.Error().Err(err).Msgf("failed to delete consumer %s:%s", gsl.streamName, gsl.consumerName)
	}
}

func (gsl *GameStateListener) ReclaimPending(ctx context.Context) error {
	for {
		msgs, _, err := gsl.rc.XAutoClaim(ctx, &redis.XAutoClaimArgs{
			Stream:   gsl.streamName,
			Group:    gsl.consumerGroupName,
			Consumer: gsl.consumerName,
			MinIdle:  10 * time.Second,
			Start:    "0-0",
			Count:    50,
		}).Result()

		if err != nil {
			return err
		}

		if len(msgs) == 0 {
			return nil
		}

		// TODO: check errors
		for _, msg := range msgs {
			_ = gsl.handleMessage(ctx, msg) 
			_ = gsl.rc.XAck(ctx, gsl.streamName, gsl.consumerGroupName, msg.ID).Err()
		}
	}
}

func (gsl *GameStateListener) handleMessage(ctx context.Context, msg redis.XMessage) error {
	rawID, ok := msg.Values["game_id"]
	if !ok {
		return common.ServiceError{
			Code:    common.ErrorRedisStream,
			Message: fmt.Sprintf("message %s has no game_id", msg.ID),
		}
	}

	var gameId int64
	switch v := rawID.(type) {
	case string:
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return common.ServiceError{
				Code:    common.ErrorRedisStream,
				Message: fmt.Sprintf("invalid game_id in message %s", msg.ID),
			}
		}
		gameId = id
	case int64:
		gameId = v
	case int:
		gameId = int64(v)
	default:
		return common.ServiceError{
			Code:    common.ErrorRedisStream,
			Message: fmt.Sprintf("unexpected type for game_id: %T", v),
		}
	}

	return gsl.gm.AdvanceGame(ctx, gameId)
}

func newConsumerName(prefix, suffix string) string {
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

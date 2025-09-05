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
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type GameUpdatePublisher interface {
	PublishGameUpdate(channel string, upd models.GameUpdate) error
}

type GameWatchdog struct {
	txm           db.TxManager
	redis         *redis.Client
	interval      time.Duration
	limit         int32
	streamName    string
	streamMaxLean int64
	enqueueDelay  time.Duration
}

func NewGameWatchdog(
	txm db.TxManager, redis *redis.Client,
	interval time.Duration, limit int32,
	streamName string, streamMaxLean int64,
	enqueueDelay time.Duration,
) *GameWatchdog {
	return &GameWatchdog{
		txm:           txm,
		redis:         redis,
		interval:      interval,
		limit:         limit,
		streamName:    streamName,
		streamMaxLean: streamMaxLean,
		enqueueDelay:  enqueueDelay,
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

func (wd *GameWatchdog) EnqueueGame(gameId int64) {
	time.AfterFunc(wd.enqueueDelay, func() {
		wd.enqueueGame(gameId)
	})
}

func (wd *GameWatchdog) enqueueGame(gameId int64) {
	time.AfterFunc(wd.enqueueDelay, func() {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelFunc()

		_, err := wd.redis.XAdd(ctx, &redis.XAddArgs{
			Stream: wd.streamName,
			MaxLen: wd.streamMaxLean,
			Approx: true,
			Values: map[string]any{
				"game_id": gameId,
			},
		}).Result()
		if err != nil {
			log.Error().Err(err).Msg("failed to enqueue game")
		}
	})
}


func (wd *GameWatchdog) checkPendingGames() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancelFunc()

	err := db.WithTx(ctx, wd.txm, func(ctx context.Context, tx pgx.Tx) error {
		q := wd.txm.Querier(tx)
		games, err := q.AdvanceGames(ctx, qg.AdvanceGamesParams{Limit: wd.limit})

		if err != nil {
			return err
		}

		if len(games) > 0 {
			_, err = wd.redis.Pipelined(ctx, func(pipe redis.Pipeliner) error {
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
	redis             *redis.Client
	manager           *GameManager
	updatePublisher   GameUpdatePublisher
	streamName        string
	consumerGroupName string
	count             int64
	block             time.Duration
	idleDLQ           time.Duration
	consumerName      string
	streamMaxLean     int64
}

func NewGameStateListener(
	redis *redis.Client, manager *GameManager, updatePublisher GameUpdatePublisher,
	streamName, consumerGroupName, consumerId string,
	count int64, block, idleDLQ time.Duration,
	streamMaxLean int64) *GameStateListener {
	return &GameStateListener{
		redis:             redis,
		manager:           manager,
		updatePublisher:   updatePublisher,
		streamName:        streamName,
		consumerGroupName: consumerGroupName,
		count:             count,
		block:             block,
		idleDLQ:           idleDLQ,
		streamMaxLean:     streamMaxLean,
		consumerName:      newConsumerName("worker", consumerId),
	}
}

func (gsl *GameStateListener) Listen(ctx context.Context) {
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		default:
			res, err := gsl.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
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
					gameId, err := parseGameId(msg)
					if err != nil {
						log.Error().Err(err).Msgf("failed to parse game id %s:%s", gsl.streamName, msg.ID)
						continue
					}

					if err := gsl.handleMessage(ctx, gameId); err != nil {
						log.Error().Err(err).Msgf("handler failed for %s:%s", gsl.streamName, msg.ID)
						continue
					}

					if err := gsl.redis.XAck(ctx, gsl.streamName, gsl.consumerGroupName, msg.ID).Err(); err != nil {
						log.Error().Err(err).Msgf("ack failed %s:%s", gsl.streamName, msg.ID)
					}
				}
			}
		}
	}

	delCtx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	err := gsl.redis.XGroupDelConsumer(delCtx, gsl.streamName, gsl.consumerGroupName, gsl.consumerName).Err()
	if err != nil {
		log.Error().Err(err).Msgf("failed to delete consumer %s:%s", gsl.streamName, gsl.consumerName)
	}
}

func (gsl *GameStateListener) ReclaimPending(ctx context.Context) error {
	for {
		msgs, _, err := gsl.redis.XAutoClaim(ctx, &redis.XAutoClaimArgs{
			Stream:   gsl.streamName,
			Group:    gsl.consumerGroupName,
			Consumer: gsl.consumerName,
			MinIdle:  gsl.idleDLQ,
			Start:    "0-0",
			Count:    50,
		}).Result()

		if err != nil {
			return err
		}

		if len(msgs) == 0 {
			return nil
		}

		for _, msg := range msgs {
			gameId, err := parseGameId(msg)
			if err != nil {
				continue
			}
			err = gsl.handleMessage(ctx, gameId)
			if err != nil {
				log.Error().Err(err).Msgf("reclaim handler failed for %s:%s", gsl.streamName, msg.ID)
			}

			err = gsl.redis.XAck(ctx, gsl.streamName, gsl.consumerGroupName, msg.ID).Err()
			if err != nil {
				log.Error().Err(err).Msgf("reclaim ack failed %s:%s", gsl.streamName, msg.ID)
			}
		}
	}
}

func (gsl *GameStateListener) handleMessage(ctx context.Context, gameId int64) error {
	upd, err := gsl.manager.AdvanceGame(ctx, gameId)
	if err != nil {
		return err
	}

	if upd != nil {
		return gsl.updatePublisher.PublishGameUpdate(GetCfChannelName(gameId), *upd)
	}

	return nil
}

func parseGameId(msg redis.XMessage) (int64, error) {
	rawID, ok := msg.Values["game_id"]
	if !ok {
		return 0, common.ServiceError{
			Code:    common.ErrorRedisStream,
			Message: fmt.Sprintf("message %s has no game_id", msg.ID),
		}
	}

	var gameId int64
	switch v := rawID.(type) {
	case string:
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, common.ServiceError{
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
		return 0, common.ServiceError{
			Code:    common.ErrorRedisStream,
			Message: fmt.Sprintf("unexpected type for game_id: %T", v),
		}
	}

	return gameId, nil
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

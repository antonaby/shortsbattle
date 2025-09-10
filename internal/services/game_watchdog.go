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
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type GameUpdatePublisher interface {
	PublishGameUpdate(channel string, upd models.GameUpdate) error
}

type GameWatchdogConfig struct {
	CronStr           string
	EnqueueInterval   time.Duration
	WriteBatchSize    int32
	ReadBatchSize     int64
	StreamName        string
	ConsumerGroupName string
	StreamMaxLean     int64
	ReadWait          time.Duration
	IdleDLQ           time.Duration
}

type GameWatchdog struct {
	txm    db.TxManager
	redis  *redis.Client
	config GameWatchdogConfig
}

func NewGameWatchdog(txm db.TxManager, redis *redis.Client, config GameWatchdogConfig) *GameWatchdog {
	return &GameWatchdog{
		txm:    txm,
		redis:  redis,
		config: config,
	}
}

// TODO: add check interval (like 5 sec, instead of fetching only when next_state exceeded)
// TODO: use github.com/go-co-op/gocron
func (wd *GameWatchdog) Run(ctx context.Context) error {
	t := time.NewTicker(1 * time.Second)
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
	wd.enqueueGame(gameId)
}

func (wd *GameWatchdog) enqueueGame(gameId int64) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	_, err := wd.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: wd.config.StreamName,
		MaxLen: wd.config.StreamMaxLean,
		Approx: true,
		Values: map[string]any{
			"game_id": gameId,
		},
	}).Result()

	if err != nil {
		log.Error().Stack().Err(err).Msg("failed to enqueue game")
	}
}

func (wd *GameWatchdog) checkPendingGames() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancelFunc()

	err := db.WithTx(ctx, wd.txm, func(ctx context.Context, tx pgx.Tx) error {
		q := wd.txm.Querier(tx)
		games, err := q.AdvanceGames(ctx, wd.config.WriteBatchSize, db.ToPgInterval(wd.config.EnqueueInterval))

		if err != nil {
			return err
		}

		if len(games) > 0 {
			_, err = wd.redis.Pipelined(ctx, func(pipe redis.Pipeliner) error {
				for _, game := range games {
					pipe.XAdd(ctx, &redis.XAddArgs{
						Stream: wd.config.StreamName,
						MaxLen: wd.config.StreamMaxLean,
						Approx: true,
						Values: map[string]any{
							"game_id": game.GameID,
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
		log.Error().Err(err).Stack().Msg("failed to advance game states")
	}
}

// TODO: add multiple workers
type GameListener struct {
	redis           *redis.Client
	manager         *GameManager
	updatePublisher GameUpdatePublisher
	config          GameWatchdogConfig
	ConsumerName    string
}

func NewGameListener(redis *redis.Client, manager *GameManager, updatePublisher GameUpdatePublisher, config GameWatchdogConfig) *GameListener {
	return &GameListener{
		redis:           redis,
		manager:         manager,
		updatePublisher: updatePublisher,
		config:          config,
		ConsumerName:    newConsumerName("worker", "1"),
	}
}

func (gsl *GameListener) EnsureStreamGroup() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	err := gsl.redis.XGroupCreateMkStream(ctx, gsl.config.StreamName, gsl.config.ConsumerGroupName, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}

	return nil
}

func (gsl *GameListener) Listen(ctx context.Context) {
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		default:
			res, err := gsl.redis.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    gsl.config.ConsumerGroupName,
				Consumer: gsl.ConsumerName,
				Streams:  []string{gsl.config.StreamName, ">"},
				Count:    gsl.config.ReadBatchSize,
				Block:    gsl.config.ReadWait,
				NoAck:    false,
			}).Result()

			if err == redis.Nil {
				continue
			}

			if err != nil {
				log.Error().Err(err).Stack().Msgf("readgroup error %s", gsl.config.StreamName)
				continue
			}

			for _, str := range res {
				for _, msg := range str.Messages {
					gameId, err := parseGameId(msg)
					if err != nil {
						log.Error().Err(err).Stack().Msgf("failed to parse game id %s:%s", gsl.config.StreamName, msg.ID)
						continue
					}

					if err := gsl.handleMessage(ctx, gameId); err != nil {
						log.Error().Err(err).Stack().Msgf("handler failed for %s:%s", gsl.config.StreamName, msg.ID)
						continue
					}

					if err := gsl.redis.XAck(ctx, gsl.config.StreamName, gsl.config.ConsumerGroupName, msg.ID).Err(); err != nil {
						log.Error().Err(err).Stack().Msgf("ack failed %s:%s", gsl.config.StreamName, msg.ID)
					}
				}
			}
		}
	}

	delCtx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	err := gsl.redis.XGroupDelConsumer(delCtx, gsl.config.StreamName, gsl.config.ConsumerGroupName, gsl.ConsumerName).Err()
	if err != nil {
		log.Error().Err(err).Stack().Msgf("failed to delete consumer %s:%s", gsl.config.StreamName, gsl.ConsumerName)
	}
}

func (gsl *GameListener) ReclaimPending(ctx context.Context) error {
	for {
		msgs, _, err := gsl.redis.XAutoClaim(ctx, &redis.XAutoClaimArgs{
			Stream:   gsl.config.StreamName,
			Group:    gsl.config.ConsumerGroupName,
			Consumer: gsl.ConsumerName,
			MinIdle:  gsl.config.IdleDLQ,
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
				log.Error().Err(err).Stack().Msgf("reclaim handler failed for %s:%s", gsl.config.StreamName, msg.ID)
			}

			err = gsl.redis.XAck(ctx, gsl.config.StreamName, gsl.config.ConsumerGroupName, msg.ID).Err()
			if err != nil {
				log.Error().Err(err).Stack().Msgf("reclaim ack failed %s:%s", gsl.config.StreamName, msg.ID)
			}
		}
	}
}

func (gsl *GameListener) handleMessage(ctx context.Context, gameId int64) error {
	// upd, err := gsl.manager.AdvanceGame(ctx, gameId)
	// if err != nil {
	// 	return err
	// }

	// if upd != nil {
	// 	return gsl.updatePublisher.PublishGameUpdate(GetCfChannelName(gameId), *upd)
	// }

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

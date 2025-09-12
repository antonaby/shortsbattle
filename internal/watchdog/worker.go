package watchdog

import (
	"context"
	"fmt"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func wdError(code common.ErrorCode, msg string, err error) error {
	return common.ServiceError{
		Code:    code,
		Message: fmt.Sprintf("watchdog: %s", msg),
		Cause:   err,
	}
}

type WatchdogWorkerLogger struct {
	log zerolog.Logger
}

func (wl WatchdogWorkerLogger) Debug(args ...any) {
	wl.log.Debug().Msgf("%v", args...)
}

func (wl WatchdogWorkerLogger) Info(args ...any) {
	wl.log.Info().Msgf("%v", args...)
}

func (wl WatchdogWorkerLogger) Warn(args ...any) {
	wl.log.Warn().Msgf("%v", args...)
}

func (wl WatchdogWorkerLogger) Error(args ...any) {
	wl.log.Error().Msgf("%v", args...)
}

func (wl WatchdogWorkerLogger) Fatal(args ...any) {
	wl.log.Fatal().Msgf("%v", args...)
}

type WatchdogWorker struct {
	server *asynq.Server
	mux    *asynq.ServeMux
}

func NewWatchdogWorker(redisHost string, advanceAtProcessor *AdvanceGameAtProcessor, advanceNowProcessor *AdvanceGameNowProcessor) *WatchdogWorker {
	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisHost},
		asynq.Config{
			Concurrency: 10,
			Logger: WatchdogWorkerLogger{
				log: log.Logger,
			},
			// TODO: add error handler and other params
		},
	)

	mux := asynq.NewServeMux()
	mux.Handle(JobTypeAdvanceAt, advanceAtProcessor)
	mux.Handle(JobTypeAdvanceNow, advanceNowProcessor)

	return &WatchdogWorker{
		server: server,
		mux:    mux,
	}
}

func (worker *WatchdogWorker) Run() error {
	if err := worker.server.Start(worker.mux); err != nil {
		return wdError(common.ErrorInit, "failed to start asynq worker", err)
	}

	return nil
}

func (worker *WatchdogWorker) Shutdown() {
	worker.server.Shutdown()
}

type WatchdogClient struct {
	client *asynq.Client
}

func NewWatchdogClient(redisHost string) *WatchdogClient {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisHost})

	return &WatchdogClient{
		client: client,
	}
}

func (publisher *WatchdogClient) Close() error {
	err := publisher.client.Close()
	if err != nil {
		return wdError(common.ErrorInit, "failed to stop asynq client", err)
	}

	return nil
}

func (publisher *WatchdogClient) AdvanceGameNow(ctx context.Context, gameId int64) error {
	task, err := NewAdvanceGameNowTask(gameId)
	if err != nil {
		return wdError(common.ErrorEnqueue, "failed to create task", err)
	}

	info, err := publisher.client.EnqueueContext(ctx, task)
	if err != nil {
		return wdError(common.ErrorEnqueue, "failed to enqueue task", err)
	}

	log.Debug().Msgf("watchdog: enqueued AdvanceGameNow task: %s", info.ID)
	return nil
}

func (publisher *WatchdogClient) AdvanceGameAt(ctx context.Context, processAt time.Time, gameId int64, updateKey uuid.UUID, isTimeout bool) error {
	task, err := NewAdvanceGameAtTask(gameId, updateKey, isTimeout)
	if err != nil {
		return wdError(common.ErrorEnqueue, "failed to create task", err)
	}

	info, err := publisher.client.EnqueueContext(ctx, task, asynq.ProcessAt(processAt))
	if err != nil {
		return wdError(common.ErrorEnqueue, "failed to enqueue task", err)
	}

	log.Debug().Msgf("watchdog: enqueued AdvanceGameAt task: %s", info.ID)
	return nil
}

type DBWatchdog struct {
	txm            db.TxManager
	updateInterval time.Duration
	client         *WatchdogClient
}

func NewDBWatchdog(txm db.TxManager, client *WatchdogClient, updateInterval time.Duration) *DBWatchdog {
	return &DBWatchdog{
		txm:            txm,
		client:         client,
		updateInterval: updateInterval,
	}
}

func (watchdog DBWatchdog) Run(ctx context.Context) error {
	t := time.NewTicker(watchdog.updateInterval)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			watchdog.enqueueGames()
		}
	}
}

func (watchdog DBWatchdog) enqueueGames() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancelFunc()

	err := db.WithTxQ(ctx, watchdog.txm, func(ctx context.Context, q qg.Querier) error {
		games, err := q.EnqueueGames(ctx)
		if err != nil {
			return wdError(common.GetDbErrorCode(err), "failed to enquque games", err)
		}

		for _, g := range games {
			if g.NextGameUpdateAt.Valid {
				err := watchdog.client.AdvanceGameAt(ctx, g.NextGameUpdateAt.Time, g.GameID, g.UpdateKey.Bytes, true)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Stack().Send()
	}
}

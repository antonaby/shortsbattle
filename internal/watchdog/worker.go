package watchdog

import (
	"context"
	"fmt"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
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

func NewWatchdogWorker(redisHost string, advanceProcessor *AdvanceGameProcessor) *WatchdogWorker {
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
	mux.Handle(JobTypeAdvance, advanceProcessor)

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

func (publisher *WatchdogClient) AdvanceGame(ctx context.Context, gameId int64, processIn time.Duration) error {
	task, err := NewAdvanceGameTask(gameId)
	if err != nil {
		return wdError(common.ErrorEnqueue, "failed to create task", err)
	}

	info, err := publisher.client.EnqueueContext(ctx, task, asynq.ProcessIn(processIn))
	if err != nil {
		return wdError(common.ErrorEnqueue, "failed to enqueue task", err)
	}

	log.Debug().Msgf("watchdog: enqueued advance game task: %s", info.ID)
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
	log.Debug().Msg("Test")
}

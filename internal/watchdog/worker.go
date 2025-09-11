package watchdog

import (
	"context"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type WatchdogLogger struct {
	log zerolog.Logger
}

func (wl WatchdogLogger) Debug(args ...any) {
	wl.log.Debug().Msgf("%v", args...)
}

func (wl WatchdogLogger) Info(args ...any) {
	wl.log.Info().Msgf("%v", args...)
}

func (wl WatchdogLogger) Warn(args ...any) {
	wl.log.Warn().Msgf("%v", args...)
}

func (wl WatchdogLogger) Error(args ...any) {
	wl.log.Error().Msgf("%v", args...)
}

func (wl WatchdogLogger) Fatal(args ...any) {
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
			Logger: WatchdogLogger{
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
		return err
	}

	return nil
}

func (worker *WatchdogWorker) Shutdown() {
	worker.server.Shutdown()
}

type WatchdogPublisher struct {
	client *asynq.Client
}

func NewWatchdogPublisher(redisHost string) *WatchdogPublisher {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisHost})

	return &WatchdogPublisher{
		client: client,
	}
}

func (publisher *WatchdogPublisher) Close() error {
	return publisher.client.Close()
}

func (publisher *WatchdogPublisher) AdvanceGame(ctx context.Context, gameId int64, processIn time.Duration) error {
	task, err := NewAdvanceGameTask(gameId)
	if err != nil {
		return err
	}

	info, err := publisher.client.EnqueueContext(ctx, task, asynq.ProcessIn(processIn))
	if err != nil {
		return common.ServiceError{
			Code:    common.ErrorEnqueue,
			Message: "failed to enqueue task",
			Cause:   err,
		}
	}

	log.Debug().Msgf("enqueued advance game task: %s", info.ID)
	return nil
}

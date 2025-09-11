package watchdog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/hibiken/asynq"
)

const (
	JobTypeAdvance = "game:advance"
)

type AdvanceGamePayload struct {
	GameID int64
}

func NewAdvanceGameTask(gameId int64) (*asynq.Task, error) {
	payload, err := json.Marshal(AdvanceGamePayload{GameID: gameId})
	if err != nil {
		return nil, wdError(common.ErrorMarshal, "failed to marshal palyload", err)
	}

	// TODO: set max retry if needed
	return asynq.NewTask(JobTypeAdvance, payload, asynq.MaxRetry(10)), nil
}

type GameManager interface {
	AdvanceGame(ctx context.Context, gameId int64) (*models.TaskReschedule, *models.GameUpdate, error)
}

type GameUpdatePublisher interface {
	PublishGameUpdate(upd models.GameUpdate) error
}

type AdvanceGameProcessor struct {
	gameManager     GameManager
	updatePublisher GameUpdatePublisher
	watchdog        *WatchdogClient
}

func NewAdvanceGameProcessor(gameManager GameManager, updatePublisher GameUpdatePublisher, watchdog *WatchdogClient) *AdvanceGameProcessor {
	return &AdvanceGameProcessor{
		gameManager:     gameManager,
		updatePublisher: updatePublisher,
		watchdog:        watchdog,
	}
}

func (processor *AdvanceGameProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload AdvanceGamePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshaling failed: %v: %w", err, asynq.SkipRetry)
	}

	res, upd, err := processor.gameManager.AdvanceGame(ctx, payload.GameID)
	if err != nil {
		return err
	}

	if res != nil {
		err = processor.watchdog.AdvanceGame(ctx, res.GameID, res.ProcessIn)
		if err != nil {
			return err
		}
	}

	if upd != nil {
		return processor.updatePublisher.PublishGameUpdate(*upd)
	}

	return nil
}

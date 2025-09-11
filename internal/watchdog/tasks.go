package watchdog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/services"
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
		return nil, common.ServiceError{
			Code: common.ErrorMarshal,
			Message: "failed to marshal palyload",
			Cause: err,
		}
	}

	return asynq.NewTask(JobTypeAdvance, payload), nil
}

type AdvanceGameProcessor struct {
	gameManager *services.GameManager
}

func NewAdvanceGameProcessor(gameManager *services.GameManager) *AdvanceGameProcessor {
	return &AdvanceGameProcessor{
		gameManager: gameManager,
	}
}

func (processor *AdvanceGameProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload AdvanceGamePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshaling failed: %v: %w", err, asynq.SkipRetry)
	}

	_, err := processor.gameManager.AdvanceGame(ctx, payload.GameID)
	return err
}

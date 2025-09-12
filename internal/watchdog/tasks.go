package watchdog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

const (
	JobTypeAdvance = "game:advance"
)

type AdvanceGamePayload struct {
	GameID    int64
	UpdateKey uuid.UUID
	IsTimeout bool
}

func NewAdvanceGameTask(gameId int64, updateKey uuid.UUID, isTimeout bool) (*asynq.Task, error) {
	payload, err := json.Marshal(AdvanceGamePayload{
		GameID:    gameId,
		UpdateKey: updateKey,
		IsTimeout: isTimeout,
	})
	if err != nil {
		return nil, wdError(common.ErrorMarshal, "failed to marshal palyload", err)
	}

	// TODO: set max retry if needed
	return asynq.NewTask(JobTypeAdvance, payload, asynq.MaxRetry(10)), nil
}

type GameManager interface {
	AdvanceGame(ctx context.Context, gameId int64, updateKey uuid.UUID, isTimeout bool) (*models.GameUpdate, error)
}

type GameUpdatePublisher interface {
	PublishGameUpdate(upd models.GameUpdate) error
}

type AdvanceGameProcessor struct {
	gameManager     GameManager
	updatePublisher GameUpdatePublisher
}

func NewAdvanceGameProcessor(gameManager GameManager, updatePublisher GameUpdatePublisher) *AdvanceGameProcessor {
	return &AdvanceGameProcessor{
		gameManager:     gameManager,
		updatePublisher: updatePublisher,
	}
}

func (processor *AdvanceGameProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload AdvanceGamePayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshaling failed: %v: %w", err, asynq.SkipRetry)
	}

	upd, err := processor.gameManager.AdvanceGame(ctx, payload.GameID, payload.UpdateKey, payload.IsTimeout)
	if err != nil {
		return err
	}

	if upd != nil {
		return processor.updatePublisher.PublishGameUpdate(*upd)
	}

	return nil
}

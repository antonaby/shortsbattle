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
	JobTypeAdvanceAt  = "game:advance_at"
	JobTypeAdvanceNow = "game:advance_now"
)

type AdvanceGameAtPayload struct {
	GameID    int64
	UpdateKey uuid.UUID
	IsTimeout bool
}

func NewAdvanceGameAtTask(gameId int64, updateKey uuid.UUID, isTimeout bool) (*asynq.Task, error) {
	payload, err := json.Marshal(AdvanceGameAtPayload{
		GameID:    gameId,
		UpdateKey: updateKey,
		IsTimeout: isTimeout,
	})
	if err != nil {
		return nil, wdError(common.ErrorMarshal, "failed to marshal palyload", err)
	}

	// TODO: set max retry if needed
	return asynq.NewTask(JobTypeAdvanceAt, payload, asynq.MaxRetry(10)), nil
}

type AdvanceGameNowPayload struct {
	GameID int64
}

func NewAdvanceGameNowTask(gameId int64) (*asynq.Task, error) {
	payload, err := json.Marshal(AdvanceGameNowPayload{
		GameID: gameId,
	})
	if err != nil {
		return nil, wdError(common.ErrorMarshal, "failed to marshal palyload", err)
	}

	// TODO: set max retry if needed
	return asynq.NewTask(JobTypeAdvanceNow, payload, asynq.MaxRetry(10)), nil
}

type GameManager interface {
	AdvanceGameNow(context.Context, int64) (*models.GameUpdate, error)
	AdvanceGameAt(context.Context, int64, uuid.UUID, bool) (*models.GameUpdate, error)
}

type GameUpdatePublisher interface {
	PublishGameUpdate(upd models.GameUpdate) error
}

type AdvanceGameAtProcessor struct {
	gameManager     GameManager
	updatePublisher GameUpdatePublisher
}

func NewAdvanceGameAtProcessor(gameManager GameManager, updatePublisher GameUpdatePublisher) *AdvanceGameAtProcessor {
	return &AdvanceGameAtProcessor{
		gameManager:     gameManager,
		updatePublisher: updatePublisher,
	}
}

func (processor *AdvanceGameAtProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload AdvanceGameAtPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshaling failed: %v: %w", err, asynq.SkipRetry)
	}

	upd, err := processor.gameManager.AdvanceGameAt(ctx, payload.GameID, payload.UpdateKey, payload.IsTimeout)
	if err != nil {
		return err
	}

	if upd != nil {
		return processor.updatePublisher.PublishGameUpdate(*upd)
	}

	return nil
}

type AdvanceGameNowProcessor struct {
	gameManager     GameManager
	updatePublisher GameUpdatePublisher
}

func NewAdvanceGameNowProcessor(gameManager GameManager, updatePublisher GameUpdatePublisher) *AdvanceGameNowProcessor {
	return &AdvanceGameNowProcessor{
		gameManager:     gameManager,
		updatePublisher: updatePublisher,
	}
}

func (processor *AdvanceGameNowProcessor) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var payload AdvanceGameNowPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshaling failed: %v: %w", err, asynq.SkipRetry)
	}

	upd, err := processor.gameManager.AdvanceGameNow(ctx, payload.GameID)
	if err != nil {
		return err
	}

	if upd != nil {
		return processor.updatePublisher.PublishGameUpdate(*upd)
	}

	return nil
}
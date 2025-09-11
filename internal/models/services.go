package models

import (
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/jackc/pgx/v5/pgtype"
)

type MessageType string

const (
	GameDetailsMsg  MessageType = "details"
	GameUpdateMsg   MessageType = "state_change"
	GameCompleteMsg MessageType = "complete"
)

type StageChangeReason string

const (
	ReasonLobbyFull    StageChangeReason = "lobby-full"
	ReasonLobbyTimeout StageChangeReason = "lobby-timeout"
)

type GameUpdate struct {
	GameID            int64              `json:"id"`
	MsgType           MessageType        `json:"msg_type"`
	Stage             qg.GameStage       `json:"state"`
	StateChangeReason *StageChangeReason `json:"state_change_reason,omitempty"`
	RoundN            int32              `json:"round"`
	StateChangedAt    pgtype.Timestamptz `json:"state_changed_at"`
	RemainingTimeMs   int64              `json:"remaning_time_ms,omitempty"`
	Theme             *qg.Theme          `json:"theme,omitempty"`
	Result            *GameResult        `json:"result,omitempty"`
}

type GameResult struct {
}

package models

import (
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/jackc/pgx/v5/pgtype"
)

type ThemeExt struct {
	ID          int64              `json:"id"`
	Name        string             `json:"name"`
	Description pgtype.Text        `json:"description"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	Requests    []qg.VideoRequest  `json:"requests"`
}

type MessageType string

const (
	GameDetailsMsg MessageType = "details"
	GameUpdateMsg  MessageType = "state_change"
)

type GameUpdate struct {
	GameID            int64              `json:"id"`
	MsgType           MessageType        `json:"msg_type"`
	State             qg.GameState       `json:"state"`
	StateChangedAt    pgtype.Timestamptz `json:"state_changed_at"`
	NextStateChangeAt pgtype.Timestamptz `json:"next_state_change_at"`
	RamaningTimeMs    int64              `json:"remaning_time_ms"`
}

type GameDetails struct {
	GameID            int64              `json:"id"`
	MsgType           MessageType        `json:"msg_type"`
	State             qg.GameState       `json:"state"`
	StateChangedAt    pgtype.Timestamptz `json:"state_changed_at"`
	NextStateChangeAt pgtype.Timestamptz `json:"next_state_change_at"`
	RamaningTimeMs    int64              `json:"remaning_time_ms"`
}

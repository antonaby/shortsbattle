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

type GameUpdate struct {
	ID                int64              `json:"id"`
	ThemeID           int64              `json:"theme_id"`
	State             qg.GameState       `json:"state"`
	CreatedAt         pgtype.Timestamptz `json:"created_at"`
	StateChangedAt    pgtype.Timestamptz `json:"state_changed_at"`
	NextStateChangeAt pgtype.Timestamptz `json:"next_state_change_at"`
}

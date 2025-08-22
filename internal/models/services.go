package models

import (
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/jackc/pgx/v5/pgtype"
)

type GameState string

const (
	StateLobby GameState = "lobby"
)

type ThemeExt struct {
	ID          int64              `json:"id"`
	Name        string             `json:"name"`
	Description pgtype.Text        `json:"description"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	Requests    []qg.VideoRequest  `json:"requests"`
}

type GemeDetails struct {
	ThemeID int64
	Status  string
}

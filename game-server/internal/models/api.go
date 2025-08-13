package models

import (
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/jackc/pgx/v5/pgtype"
)

type OkResponse struct {
	Msg string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type AddPlayerToGame struct {
	PlayerId int64 `json:"player_id" validate:"required"`
}

type CreateGame struct {
	ThemeId int64 `json:"theme_id" validate:"required"`
}

type CreateThemeParams struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type GameDetails struct {
	ID                 int64            `json:"id"`
	ThemeID            int64            `json:"theme_id"`
	Status             db.GameStatus    `json:"status"`
	CreatedAt          pgtype.Timestamp `json:"created_at"`
	StageTimeRemaining time.Duration    `json:"stage_time_remaining"`
}

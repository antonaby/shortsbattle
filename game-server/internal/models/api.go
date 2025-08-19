package models

import (
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

type SubmitVideoToGame struct {
	PlayerId int64  `json:"player_id" validate:"required"`
	VideoUrl string `json:"video_url" validate:"required"`
}

type SubmitVoteToGame struct {
	GameID  int64        `json:"game_id" validate:"required"`
	VideoID int64        `json:"video_id" validate:"required"`
	VoterID int64        `json:"voter_id" validate:"required"`
	Value   db.VoteValue `json:"value" validate:"required"`
}

type CreateGame struct {
	ThemeId  int64 `json:"theme_id" validate:"required"`
	PlayerId int64 `json:"player_id" validate:"required"`
}

type CreateThemeParams struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type GameDetails struct {
	ID                 int64              `json:"id"`
	Theme              db.Theme           `json:"theme"`
	Status             db.GameStatus      `json:"status"`
	CreatedAt          pgtype.Timestamptz `json:"created_at"`
	StageTimeRemaining int64              `json:"stage_time_remaining"`
	Players            []db.Player        `json:"players"`
	Videos             []db.Video         `json:"videos"`
	Votes              []db.Vote          `json:"votes"`
}

type GameUpdate struct {
	ID                 int64         `json:"id"`
	Status             db.GameStatus `json:"status"`
	StageTimeRemaining int64         `json:"stage_time_remaining"`
}

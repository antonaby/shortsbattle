package models

import (
	"encoding/json"

	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
)

type OkResponse struct {
	Msg string `json:"message"`
}

type OkGameIdReposne struct {
	Msg    string `json:"message"`
	GameId int64  `json:"game_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateThemeRequest struct {
	Title       string      `json:"title" validate:"required,min=3"`
	Description *string     `json:"description"`
	Mode        qg.GameMode `json:"mode" validate:"required"`
}

type CreateRoundRequest struct {
	RoundN      int32   `json:"round_n" validate:"required,gt=0"`
	Title       string  `json:"title" validate:"required,min=3"`
	Description *string `json:"description"`
}

type AddVideoRequest struct {
	VideoUrl string `json:"video_url" validate:"required,url"`
}

type SubmitVideoRequest struct {
	VideoID  *int64  `json:"video_id"`
	VideoUrl *string `json:"video_url" validate:"omitempty,url"`
}

type UpdateGameModeRequest struct {
	Mode qg.PlayerGameMode `json:"mode" validate:"required"`
}

type JoinGameRequest struct {
	ThemeID int64             `json:"theme_id" validate:"required"`
	Mode    qg.PlayerGameMode `json:"mode" validate:"required"`
}

type VoteForVideoRequest struct {
	Value json.RawMessage `json:"value" validate:"required"`
}

type VideoErrRequest struct {
	ErrMsg string `json:"error_msg" validate:"required"`
}

type TgUserAuthRequest struct {
	InitData string `json:"init_data" validate:"required"`
}

type OkTgUserTokenResponse struct {
	Token string `json:"token"`
}

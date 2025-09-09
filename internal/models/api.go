package models

import (
	"encoding/json"
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
	Title       string  `json:"title" validate:"required,min=3"`
	Description *string `json:"description"`
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

type JoinGameRequest struct {
	ThemeID int64 `json:"theme_id" validate:"required"`
}

type VoteForVideoRequest struct {
	Value json.RawMessage `json:"value" validate:"required"`
}

type TgUserAuthRequest struct {
	InitData string `json:"init_data" validate:"required"`
}

type OkTgUserTokenResponse struct {
	Token string `json:"token"`
}

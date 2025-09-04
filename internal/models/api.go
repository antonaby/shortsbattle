package models

import "github.com/antonaby/shortsbattle/game-server/internal/db/qg"

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
	Name        string  `json:"name" validate:"required,min=3"`
	Description *string `json:"description"`
}

type CreateVideoRequest struct {
	Request string `json:"request" validate:"required,min=3"`
}

type AddVideoRequest struct {
	VideoUrl string `json:"video_url" validate:"required,url"`
}

type SubmitVideoRequest struct {
	VideoRequestId int64   `json:"video_request_id" validate:"required"`
	VideoID        *int64  `json:"video_id"`
	VideoUrl       *string `json:"video_url" validate:"omitempty,url"`
}

type JoinGameRequest struct {
	ThemeID int64 `json:"theme_id" validate:"required"`
}

type VoteForVideoRequest struct {
	VideoID int64        `json:"video_id" validate:"required"`
	Value   qg.VoteValue `json:"value" validate:"required"`
}

type TgUserAuthRequest struct {
	InitData string `json:"init_data" validate:"required"`
}

type OkTgUserTokenResponse struct {
	Token string `json:"token"`
}

package models

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
	PlayerID int64  `json:"player_id" validate:"required"`
	VideoUrl string `json:"video_url" validate:"required,url"`
}

type SubmitVideoRequest struct {
	VideoID  *int64  `json:"video_id"`
	VideoUrl *string `json:"video_url" validate:"omitempty,url"`
	PlayerID int64   `json:"player_id" validate:"required"`
}

type JoinGameRequest struct {
	ThemeID  int64 `json:"theme_id" validate:"required"`
	PlayerID int64 `json:"player_id" validate:"required"`
}

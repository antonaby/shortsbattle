package models

type OkResponse struct {
	Msg string `json:"message"`
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

type SubmitVideoRequest struct {
	PlayerID int64  `json:"player_id" validate:"required"`
	VideoUrl string `json:"video_url" validate:"required,url"`
}

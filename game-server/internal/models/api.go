package models

type OkResponse struct {
	Msg string `json:"message"`
}

type ErrorResponse struct {
  Error string `json:"error"`
}

type CreatePlayerRequest struct {
	Username string `json:"username" validate:"required"`
}

type AddPlayerToGame struct {
	PlayerId int64 `json:"player_id" validate:"required"`
}
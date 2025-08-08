package models

type OkResponse struct {
	Msg string `json:"message"`
}

type ErrorResponse struct {
  Error string `json:"error"`
}

type CreatePlayerRequest struct {
	Username string  `json:"username" validate:"required"`
}
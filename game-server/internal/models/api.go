package models

type ErrorResponse struct {
    Error string `json:"error"`
}

type CreatePlayerRequest struct {
	Username string  `json:"username" validate:"required"`
}
package models

type OkResponse struct {
	Msg string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateThemeRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type CreateVideoRequest struct {
	Request string `json:"request"`
}

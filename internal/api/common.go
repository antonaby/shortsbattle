package api

import (
	"strconv"

	"github.com/go-playground/validator/v10"
)

const (
	InvalidIdFormatMsg        = "invalid id format"
	InvalidRequestFormatMsg   = "invalid request format"
	RequestValidationErrorMsg = "request validation error"
)

type CustomValidator struct {
	validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func parseInt64(str string) (int64, error) {
	value, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}

	return value, nil
}

package api

import (
	"strconv"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

const (
	InvalidIdFormatMsg        = "invalid id format"
	InvalidRequestFormatMsg   = "invalid request format"
	RequestValidationErrorMsg = "request validation error"
	UnathorizedErrorMsg       = "invalid credentials"
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

func parseInt32(str string) (int32, error) {
	value, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, err
	}

	return int32(value), nil
}

func getJwt(c echo.Context) (jwt.Token, error) {
	token, ok := c.Get(JWTContextKey).(jwt.Token)
	if !ok {
		return nil, common.ServiceError{
			Code:    common.ErrorJWT,
			Message: "no token in context",
		}
	}

	return token, nil
}

func getTgUserIdFromToken(c echo.Context) (int64, error) {
	token, err := getJwt(c)
	if err != nil {
		return 0, err
	}

	sub, ok := token.Subject()
	if !ok {
		return 0, common.ServiceError{
			Code:    common.ErrorJWT,
			Message: "no subject in JWT",
		}
	}

	tgId, err := parseInt64(sub)
	if err != nil {
		return 0, err
	}

	return tgId, nil
}

package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

const (
	InvalidIdFormatMsg        = "invalid id format"
	InvalidRequestFormatMsg   = "invalid request format"
	RequestValidationErrorMsg = "request validation error"
	UnathorizedErrorMsg       = "invalid credentials"
	GameActionForbiddenMsg    = "player not in game or game is in the wrong stage"
	ResourceNotFoundMsg       = "video or game not found"
	SomethingWentWrongMsg     = "something went wrong"
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

func queryParam32(c echo.Context, name string) (int32, error) {
	param, err := parseInt32(c.QueryParam(name))
	if err != nil {
		return 0, err
	}

	return param, nil
}

func param64(c echo.Context, name string) (int64, error) {
	param, err := parseInt64(c.Param(name))
	if err != nil {
		return 0, err
	}

	return param, nil
}

func sendInvalidId(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, models.ErrorResponse{
		Error: InvalidIdFormatMsg,
	})
}

func tgId(c echo.Context) (int64, error) {
	tgId, err := getTgUserIdFromToken(c)
	if err != nil {
		return 0, err
	}

	return tgId, nil
}

func sendUnauthorized(c echo.Context) error {
	return c.JSON(http.StatusUnauthorized, models.ErrorResponse{
		Error: UnathorizedErrorMsg,
	})
}

func bindAndValidate[T any](c echo.Context) (T, error) {
	request := new(T)
	if err := c.Bind(request); err != nil {
		return *request, err
	}

	if err := c.Validate(request); err != nil {
		return *request, err
	}

	return *request, nil
}

func sendInvalidReq(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, models.ErrorResponse{
		Error: InvalidRequestFormatMsg,
	})
}

func sendNotFound(c echo.Context, msg string) error {
	return c.JSON(http.StatusNotFound, models.ErrorResponse{
		Error: fmt.Sprintf("%s not found", msg),
	})
}

func logAndSendUnknowError(c echo.Context, err error) error {
	c.Echo().Logger.Errorf("request processing error: %w", err)
	return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
		Error: "Something went wrong",
	})
}

func isServErr(err error) (*common.ServiceError, bool) {
	var sErr common.ServiceError
	if errors.As(err, &sErr) {
		return &sErr, true
	}

	return nil, false
}

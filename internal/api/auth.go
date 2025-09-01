package api

import (
	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/labstack/echo/v4"
)

func (api *HttpApi) getTokenForTgUser(c echo.Context) error {
	request := new(models.TgUserAuthRequest)
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: InvalidRequestFormatMsg,
		})
	}

	if err := c.Validate(request); err != nil {
		return c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: RequestValidationErrorMsg,
		})
	}

	token, err := api.auth.NewTokenFromTgInitData(request.InitData)
	if err != nil {
		c.Echo().Logger.Errorf("failed to create game: %v", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "something went wrong",
		})
	}

	return c.JSON(http.StatusOK, models.OkTgUserTokenResponse{
		Token: string(token),
	})
}

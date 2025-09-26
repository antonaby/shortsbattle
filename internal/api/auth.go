package api

import (
	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/labstack/echo/v4"
)

func (api *HttpApi) getTokenForTgUser(c echo.Context) error {
	request, err := bindAndValidate[models.TgUserAuthRequest](c)
	if err != nil {
		return sendInvalidReq(c)
	}

	ctx := c.Request().Context()
	token, err := api.auth.NewTokenFromTgInitData(ctx, request.InitData)
	if err != nil {
		if sErr, ok := isServErr(err); ok {
			if sErr.Code == common.ErrorTgInitData {
				return sendInvalidReq(c)
			}
		}

		return logAndSendUnknowError(c, err)
	}

	return c.JSON(http.StatusOK, models.OkTgUserTokenResponse{
		Token: string(token),
	})
}

package api

import (
	"net/http"
	"strings"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/labstack/echo/v4"
)

func (api *HttpApi) addVideo(c echo.Context) error {
	tgId, err := tgId(c)
	if err != nil {
		return sendUnauthorized(c)
	}

	request, err := bindAndValidate[models.AddVideoRequest](c)
	if err != nil {
		return sendInvalidReq(c)
	}

	ctx := c.Request().Context()
	video, err := api.videos.AddVideo(ctx, tgId, request.VideoUrl)

	if err != nil {
		if sErr, ok := isServErr(err); ok {
			if sErr.Code == common.ErrorOEmbedFailed {
				return sendInvalidReq(c)
			}
		}

		return logAndSendUnknowError(c, err)
	}

	return c.JSON(http.StatusOK, video)
}

func (api *HttpApi) getVideosForPlayer(c echo.Context) error {
	tgId, err := tgId(c)
	if err != nil {
		return sendUnauthorized(c)
	}

	query := c.QueryParam("query")
	ctx := c.Request().Context()
	videos, err := api.videos.GetVideosByPlayer(ctx, tgId, strings.TrimSpace(query))
	if err != nil {
		return logAndSendUnknowError(c, err)
	}

	return c.JSON(http.StatusOK, videos)
}

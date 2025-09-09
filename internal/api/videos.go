package api

import (
	"errors"
	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	m "github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/labstack/echo/v4"
)

func (api *HttpApi) addVideo(c echo.Context) error {
	tgId, err := getTgUserIdFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, m.ErrorResponse{
			Error: UnathorizedErrorMsg,
		})
	}

	request := new(m.AddVideoRequest)
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: InvalidRequestFormatMsg,
		})
	}

	if err := c.Validate(request); err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: RequestValidationErrorMsg,
		})
	}

	ctx := c.Request().Context()
	video, err := api.videos.AddVideo(ctx, tgId, request.VideoUrl)

	if err != nil {
		var sErr common.ServiceError
		if errors.As(err, &sErr) {
			if sErr.Code == common.ErrorOEmbedFailed {
				return c.JSON(http.StatusBadRequest, m.ErrorResponse{
					Error: "wrong video url",
				})
			}
		}

		c.Echo().Logger.Errorf("failed to add video: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, video)
}

func (api *HttpApi) getVideosForPlayer(c echo.Context) error {
	tgId, err := getTgUserIdFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, m.ErrorResponse{
			Error: UnathorizedErrorMsg,
		})
	}

	ctx := c.Request().Context()
	videos, err := api.videos.GetVideosByPlayer(ctx, tgId)
	if err != nil {
		c.Echo().Logger.Errorf("failed to fetch video: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, videos)
}

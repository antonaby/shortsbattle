package api

import (
	"errors"
	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	m "github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/antonaby/shortsbattle/game-server/internal/services"
	"github.com/labstack/echo/v4"
)

type VideosApi struct {
	vs *services.VideosService
}

func NewVideosApi(vs *services.VideosService, g *echo.Group) *VideosApi {
	api := &VideosApi{
		vs: vs,
	}

	api.register(g)

	return api
}

func (api *VideosApi) register(g *echo.Group) {
	v1group := g.Group("/v1")

	v1group.POST("/videos", api.addVideo)
	v1group.GET("/videos/:id", api.getVideo)

	v1group.GET("/players/:id/videos", api.getVideosForPlayer)
}

func (api *VideosApi) addVideo(c echo.Context) error {
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
	video, err := api.vs.AddVideo(ctx, *request)

	if err != nil {
		var sErr common.ServiceError
		if errors.As(err, &sErr) {
			if sErr.Code == common.ErrorDbConstraintViolation {
				return c.JSON(http.StatusNotFound, m.ErrorResponse{
					Error: "player not found",
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

func (api *VideosApi) getVideo(c echo.Context) error {
	videoId, err := parseInt64(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: InvalidIdFormatMsg,
		})
	}

	ctx := c.Request().Context()
	video, err := api.vs.GetVideo(ctx, videoId)
	if err != nil {
		var sErr common.ServiceError
		if errors.As(err, &sErr) {
			if sErr.Code == common.ErrorDbNotFound {
				return c.JSON(http.StatusNotFound, m.ErrorResponse{
					Error: "video not found",
				})
			}
		}

		c.Echo().Logger.Errorf("failed to fetch video: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, video)
}

func (api *VideosApi) getVideosForPlayer(c echo.Context) error {
	playerId, err := parseInt64(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: InvalidIdFormatMsg,
		})
	}

	ctx := c.Request().Context()
	videos, err := api.vs.GetVideosByPlayer(ctx, playerId)
	if err != nil {
		c.Echo().Logger.Errorf("failed to fetch video: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, videos)
}

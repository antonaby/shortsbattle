package api

import (
	"errors"
	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
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

	v1group.POST("/videos", api.submitVideo)
}

func (api *VideosApi) submitVideo(c echo.Context) error {
	request := new(m.SubmitVideoRequest)
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
	video, err := api.vs.CreateVideo(ctx, qg.CreateVideoParams{
		PlayerID: request.PlayerID,
		VideoUrl: request.VideoUrl,
	})

	if err != nil {
		var sErr common.ServiceError
		if errors.As(err, &sErr) {
			if sErr.Code == common.ErrorConstraintViolation {
				return c.JSON(http.StatusNotFound, m.ErrorResponse{
					Error: "player not found",
				})
			}
		}

		c.Echo().Logger.Errorf("failed to submit video: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, video)
}

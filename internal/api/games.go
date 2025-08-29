package api

import (
	"errors"
	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	m "github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/antonaby/shortsbattle/game-server/internal/services"

	"github.com/labstack/echo/v4"
)

type GamesApi struct {
	gm *services.GameManager
}

func NewGamesApi(gm *services.GameManager, g *echo.Group) *GamesApi {
	api := &GamesApi{
		gm: gm,
	}

	api.register(g)

	return api
}

func (api *GamesApi) register(g *echo.Group) {
	v1group := g.Group("/v1")

	v1group.PUT("/games/join", api.joinGame)
	v1group.PUT("/games/:id/submit", api.submitVideo) 
	v1group.GET("/games/:id/videos", api.getVideosForGame) // TODO: return DTOs intead of DB Models
}

func (api *GamesApi) joinGame(c echo.Context) error {
	request := new(m.JoinGameRequest)
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
	gameId, err := api.gm.JoinGame(ctx, request.ThemeID, request.PlayerID)

	if err != nil {
		var sErr common.ServiceError
		if errors.As(err, &sErr) {
			if sErr.Code == common.ErrorDbNotFound {
				return c.JSON(http.StatusNotFound, m.ErrorResponse{
					Error: "theme or player not found",
				})
			}
		}

		c.Echo().Logger.Errorf("failed to create game: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "something went wrong",
		})
	}

	return c.JSON(http.StatusOK, m.OkGameIdReposne{
		Msg:    "game created",
		GameId: gameId,
	})
}

// TODO: get user id from auth data
func (api *GamesApi) submitVideo(c echo.Context) error {
	gameId, err := parseInt64(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: InvalidIdFormatMsg,
		})
	}

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
	if request.VideoID != nil {
		video, err := api.gm.SubmitExistingVideo(ctx, gameId, *request.VideoID, request.PlayerID)
		if err != nil {
			var sErr common.ServiceError
			if errors.As(err, &sErr) {
				if sErr.Code == common.ErrorDbNotFound {
					return c.JSON(http.StatusNotFound, m.ErrorResponse{
						Error: "player not in the game, or game, player or video not found",
					})
				}
			}

			c.Echo().Logger.Errorf("failed to submit video: %v", err)
			return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
				Error: "Something went wrong",
			})
		}

		return c.JSON(http.StatusOK, video)
	} else if request.VideoUrl != nil {
		video, err := api.gm.SubmitNewVideo(ctx, gameId, *request.VideoUrl, request.PlayerID)
		if err != nil {
			var sErr common.ServiceError
			if errors.As(err, &sErr) {
				if sErr.Code == common.ErrorDbNotFound {
					return c.JSON(http.StatusNotFound, m.ErrorResponse{
						Error: "player not in the game, or game or player not found",
					})
				}
				if sErr.Code == common.ErrorOEmbedFailed {
					return c.JSON(http.StatusBadRequest, m.ErrorResponse{
						Error: "wrong video url",
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

	return c.JSON(http.StatusBadRequest, m.ErrorResponse{
		Error: "video_id or video_url must be provided",
	})
}

// TODO: get user id from auth data
func (api *GamesApi) getVideosForGame(c echo.Context) error {
	gameId, err := parseInt64(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: InvalidIdFormatMsg,
		})
	}

	playerId, err := parseInt64(c.QueryParam("player_id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: InvalidIdFormatMsg,
		})
	}
	
	ctx := c.Request().Context()
	videos, err := api.gm.GetVideosToWatch(ctx, gameId, playerId)
	if err != nil {
		c.Echo().Logger.Errorf("failed to submit video: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, videos)
}

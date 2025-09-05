package api

import (
	"errors"
	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	m "github.com/antonaby/shortsbattle/game-server/internal/models"

	"github.com/labstack/echo/v4"
)

func (api *HttpApi) joinGame(c echo.Context) error {
	tgId, err := getTgUserIdFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, m.ErrorResponse{
			Error: UnathorizedErrorMsg,
		})
	}

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
	gameId, err := api.gm.JoinGame(ctx, request.ThemeID, tgId)

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

func (api *HttpApi) submitVideo(c echo.Context) error {
	tgId, err := getTgUserIdFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, m.ErrorResponse{
			Error: UnathorizedErrorMsg,
		})
	}

	roundN, err := parseInt32(c.QueryParam("round"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: InvalidIdFormatMsg,
		})
	}

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
		video, err := api.gm.SubmitExistingVideo(ctx, gameId, *request.VideoID, tgId, roundN)
		if err != nil {
			var sErr common.ServiceError
			if errors.As(err, &sErr) {
				if sErr.Code == common.ErrorDbNotFound {
					return c.JSON(http.StatusNotFound, m.ErrorResponse{
						Error: "player not in the game, or game, player or video not found",
					})
				}
				if sErr.Code == common.ErrorWrongGameState {
					return c.JSON(http.StatusBadRequest, m.ErrorResponse{
						Error: "wrong game state",
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
		video, err := api.gm.SubmitNewVideo(ctx, gameId, *request.VideoUrl, tgId, roundN)
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
				if sErr.Code == common.ErrorWrongGameState {
					return c.JSON(http.StatusBadRequest, m.ErrorResponse{
						Error: "game closed",
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

func (api *HttpApi) getVideosForGame(c echo.Context) error {
	gameId, err := parseInt64(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: InvalidIdFormatMsg,
		})
	}

	roundId, err := parseInt32(c.QueryParam("round"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: InvalidIdFormatMsg,
		})
	}

	tgId, err := getTgUserIdFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, m.ErrorResponse{
			Error: UnathorizedErrorMsg,
		})
	}

	ctx := c.Request().Context()
	videos, err := api.gm.GetVideosToWatch(ctx, gameId, tgId, roundId)
	if err != nil {
		var sErr common.ServiceError
		if errors.As(err, &sErr) {
			if sErr.Code == common.ErrorWrongGameState {
				return c.JSON(http.StatusBadRequest, m.ErrorResponse{
					Error: "wrong game state",
				})
			}
		}

		c.Echo().Logger.Errorf("failed to submit video: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, videos)
}

func (api *HttpApi) voteForVideo(c echo.Context) error {
	gameVideoId, err := parseInt64(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: InvalidIdFormatMsg,
		})
	}

	tgId, err := getTgUserIdFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, m.ErrorResponse{
			Error: UnathorizedErrorMsg,
		})
	}

	request := new(m.VoteForVideoRequest)
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
	vote, err := api.gm.VoteForVideo(ctx, qg.VoteForVideoParams{
		GameVideoID: gameVideoId,
		PlayerID:    tgId,
		Value:       request.Value,
	})

	if err != nil {
		var sErr common.ServiceError
		if errors.As(err, &sErr) {
			if sErr.Code == common.ErrorDbNotFound {
				return c.JSON(http.StatusNotFound, m.ErrorResponse{
					Error: "player not in the game, or game, player or video not found",
				})
			}
			if sErr.Code == common.ErrorDbData {
				return c.JSON(http.StatusBadRequest, m.ErrorResponse{
					Error: "invalid vote value",
				})
			}
		}

		c.Echo().Logger.Errorf("failed to vote for video: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, vote)
}

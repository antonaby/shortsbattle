package api

import (
	"errors"
	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/labstack/echo/v4"
)

func (api *HttpApi) joinGame(c echo.Context) error {
	tgId, err := tgId(c)
	if err != nil {
		return err
	}

	request, err := bindAndValidate[models.JoinGameRequest](c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	gameId, err := api.games.JoinGame(ctx, request.ThemeID, tgId, request.Mode)

	if err != nil {
		var sErr common.ServiceError
		if errors.As(err, &sErr) {
			if sErr.Code == common.ErrorNotFound {
				return c.JSON(http.StatusNotFound, models.ErrorResponse{
					Error: "theme not found",
				})
			}
		}

		c.Echo().Logger.Errorf("failed to join game: %w", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: SomethingWentWrongMsg,
		})
	}

	err = api.watchdog.AdvanceGameNow(ctx, gameId)
	if err != nil {
		c.Echo().Logger.Errorf("failed to schedule game advancing: %w", err)
	}

	return c.JSON(http.StatusOK, models.OkGameIdReposne{
		Msg:    "game joined",
		GameId: gameId,
	})
}

func (api *HttpApi) updateGameMode(c echo.Context) error {
	tgId, err := tgId(c)
	if err != nil {
		return err
	}

	gameId, err := param64(c, "id")
	if err != nil {
		return err
	}

	request, err := bindAndValidate[models.UpdateGameModeRequest](c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()

	gp, err := api.games.UpdateGameMode(ctx, gameId, tgId, request.Mode)
	if err != nil {
		var sErr common.ServiceError
		if errors.As(err, &sErr) {
			if sErr.Code == common.ErrorNotFound {
				return c.JSON(http.StatusNotFound, models.ErrorResponse{
					Error: ResourceNotFoundMsg,
				})
			}
			if sErr.Code == common.ErrorForbidden {
				return c.JSON(http.StatusForbidden, models.ErrorResponse{
					Error: GameActionForbiddenMsg,
				})
			}
		}

		c.Echo().Logger.Errorf("failed to udpate game mode: %w", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: SomethingWentWrongMsg,
		})
	}

	return c.JSON(http.StatusOK, gp)
}

func (api *HttpApi) submitVideo(c echo.Context) error {
	tgId, err := tgId(c)
	if err != nil {
		return err
	}

	roundN, err := queryParam32(c, "round")
	if err != nil {
		return err
	}

	gameId, err := param64(c, "id")
	if err != nil {
		return err
	}

	request, err := bindAndValidate[models.SubmitVideoRequest](c)
	if err != nil {
		return err
	}

	if request.VideoID != nil {
		return api.submitExistingVideo(c, gameId, *request.VideoID, tgId, roundN)
	} else if request.VideoUrl != nil {
		return api.submitNewVideo(c, gameId, *request.VideoUrl, tgId, roundN)
	}

	return c.JSON(http.StatusBadRequest, models.ErrorResponse{
		Error: "video_id or video_url must be provided",
	})
}

func (api *HttpApi) submitExistingVideo(c echo.Context, gameId, videoId, playerId int64, roundN int32) error {
	ctx := c.Request().Context()
	game, video, err := api.games.SubmitExistingVideo(ctx, gameId, videoId, playerId, roundN)
	if err != nil {
		return handleVideoSubmissionError(c, err)
	}

	err = api.watchdog.AdvanceGameNow(ctx, game.GameID)
	if err != nil {
		c.Echo().Logger.Errorf("failed to schedule game advancing: %w", err)
	}
	return c.JSON(http.StatusOK, video)
}

func (api *HttpApi) submitNewVideo(c echo.Context, gameId int64, videoUrl string, playerId int64, roundN int32) error {
	ctx := c.Request().Context()
	game, video, err := api.games.SubmitNewVideo(ctx, gameId, videoUrl, playerId, roundN)
	if err != nil {
		return handleVideoSubmissionError(c, err)
	}

	err = api.watchdog.AdvanceGameNow(ctx, game.GameID)
	if err != nil {
		c.Echo().Logger.Errorf("failed to schedule game advancing: %w", err)
	}
	return c.JSON(http.StatusOK, video)
}

func handleVideoSubmissionError(c echo.Context, err error) error {
	var sErr common.ServiceError
	if errors.As(err, &sErr) {
		if sErr.Code == common.ErrorNotFound {
			return c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error: ResourceNotFoundMsg,
			})
		}
		if sErr.Code == common.ErrorOEmbedFailed {
			return c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error: "bad video url",
			})
		}
		if sErr.Code == common.ErrorForbidden {
			return c.JSON(http.StatusForbidden, models.ErrorResponse{
				Error: GameActionForbiddenMsg,
			})
		}
	}

	c.Echo().Logger.Errorf("failed to submit video: %w", err)
	return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
		Error: SomethingWentWrongMsg,
	})
}

func (api *HttpApi) getVideosForGame(c echo.Context) error {
	gameId, err := param64(c, "id")
	if err != nil {
		return err
	}

	roundN, err := queryParam32(c, "round")
	if err != nil {
		return err
	}

	tgId, err := tgId(c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	videos, err := api.games.GetVideosToWatch(ctx, gameId, tgId, roundN)
	if err != nil {
		var sErr common.ServiceError
		if errors.As(err, &sErr) {
			if sErr.Code == common.ErrorForbidden {
				return c.JSON(http.StatusForbidden, models.ErrorResponse{
					Error: GameActionForbiddenMsg,
				})
			}
		}

		c.Echo().Logger.Errorf("failed to submit video: %w", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: SomethingWentWrongMsg,
		})
	}

	return c.JSON(http.StatusOK, videos)
}

func (api *HttpApi) voteForVideo(c echo.Context) error {
	gameVideoId, err := param64(c, "id")
	if err != nil {
		return err
	}

	tgId, err := tgId(c)
	if err != nil {
		return err
	}

	request, err := bindAndValidate[models.VoteForVideoRequest](c)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	game, vote, err := api.games.VoteForVideo(ctx, gameVideoId, tgId, request.Value)

	if err != nil {
		var sErr common.ServiceError
		if errors.As(err, &sErr) {
			if sErr.Code == common.ErrorForbidden {
				return c.JSON(http.StatusForbidden, models.ErrorResponse{
					Error: GameActionForbiddenMsg,
				})
			}
			if sErr.Code == common.ErrorBadData {
				return c.JSON(http.StatusBadRequest, models.ErrorResponse{
					Error: "invalid vote value",
				})
			}
			if sErr.Code == common.ErrorNotFound {
				return c.JSON(http.StatusNotFound, models.ErrorResponse{
					Error: "game video not found",
				})
			}
		}

		c.Echo().Logger.Errorf("failed to vote for video: %w", err)
		return c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: SomethingWentWrongMsg,
		})
	}

	err = api.watchdog.AdvanceGameNow(ctx, game.GameID)
	if err != nil {
		c.Echo().Logger.Errorf("failed to schedule game advancing: %w", err)
	}
	return c.JSON(http.StatusOK, vote)
}

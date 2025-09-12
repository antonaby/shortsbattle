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

	//api.watchdog.AdvanceGame(ctx, gameId)

	return c.JSON(http.StatusOK, models.OkGameIdReposne{
		Msg:    "game joined",
		GameId: gameId,
	})
}

// func (api *HttpApi) submitVideo(c echo.Context) error {
// 	tgId, err := tgId(c)
// 	if err != nil {
// 		return err
// 	}

// 	roundN, err := queryParam32(c, "round")
// 	if err != nil {
// 		return err
// 	}

// 	gameId, err := param64(c, "id")
// 	if err != nil {
// 		return err
// 	}

// 	request, err := bindAndValidate[m.SubmitVideoRequest](c)
// 	if err != nil {
// 		return err
// 	}

// 	if request.VideoID != nil {
// 		return api.submitExistingVideo(c, gameId, *request.VideoID, tgId, roundN)
// 	} else if request.VideoUrl != nil {
// 		return api.submitNewVideo(c, gameId, *request.VideoUrl, tgId, roundN)
// 	}

// 	return c.JSON(http.StatusBadRequest, m.ErrorResponse{
// 		Error: "video_id or video_url must be provided",
// 	})
// }

// func (api *HttpApi) submitExistingVideo(c echo.Context, gameId, videoId, playerId int64, roundN int32) error {
// 	ctx := c.Request().Context()
// 	game, video, err := api.games.SubmitExistingVideo(ctx, gameId, videoId, playerId, roundN)
// 	if err != nil {
// 		return handleVideoSubmissionError(c, err)
// 	}

// 	api.watchdog.EnqueueGame(game.ID)
// 	return c.JSON(http.StatusOK, video)
// }

// func (api *HttpApi) submitNewVideo(c echo.Context, gameId int64, videoUrl string, playerId int64, roundN int32) error {
// 	ctx := c.Request().Context()
// 	game, video, err := api.games.SubmitNewVideo(ctx, gameId, videoUrl, playerId, roundN)
// 	if err != nil {
// 		return handleVideoSubmissionError(c, err)
// 	}

// 	api.watchdog.EnqueueGame(game.ID)
// 	return c.JSON(http.StatusOK, video)
// }

// func handleVideoSubmissionError(c echo.Context, err error) error {
// 	var sErr common.ServiceError
// 	if errors.As(err, &sErr) {
// 		if sErr.Code == common.ErrorDbNotFound {
// 			return c.JSON(http.StatusNotFound, m.ErrorResponse{
// 				Error: ResourceNotFoundMsg,
// 			})
// 		}
// 		if sErr.Code == common.ErrorOEmbedFailed {
// 			return c.JSON(http.StatusBadRequest, m.ErrorResponse{
// 				Error: "bad video url",
// 			})
// 		}
// 		if sErr.Code == common.ErrorForbidden {
// 			return c.JSON(http.StatusForbidden, m.ErrorResponse{
// 				Error: GameActionForbiddenMsg,
// 			})
// 		}
// 	}

// 	c.Echo().Logger.Errorf("failed to submit video: %w", err)
// 	return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
// 		Error: SomethingWentWrongMsg,
// 	})
// }

// func (api *HttpApi) getVideosForGame(c echo.Context) error {
// 	gameId, err := param64(c, "id")
// 	if err != nil {
// 		return err
// 	}

// 	roundN, err := queryParam32(c, "round")
// 	if err != nil {
// 		return err
// 	}

// 	tgId, err := tgId(c)
// 	if err != nil {
// 		return err
// 	}

// 	ctx := c.Request().Context()
// 	videos, err := api.games.GetVideosToWatch(ctx, gameId, tgId, roundN)
// 	if err != nil {
// 		var sErr common.ServiceError
// 		if errors.As(err, &sErr) {
// 			if sErr.Code == common.ErrorForbidden {
// 				return c.JSON(http.StatusForbidden, m.ErrorResponse{
// 					Error: GameActionForbiddenMsg,
// 				})
// 			}
// 		}

// 		c.Echo().Logger.Errorf("failed to submit video: %w", err)
// 		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
// 			Error: SomethingWentWrongMsg,
// 		})
// 	}

// 	return c.JSON(http.StatusOK, videos)
// }

// func (api *HttpApi) voteForVideo(c echo.Context) error {
// 	gameVideoId, err := param64(c, "id")
// 	if err != nil {
// 		return err
// 	}

// 	tgId, err := tgId(c)
// 	if err != nil {
// 		return err
// 	}

// 	request, err := bindAndValidate[m.VoteForVideoRequest](c)
// 	if err != nil {
// 		return err
// 	}

// 	ctx := c.Request().Context()
// 	game, vote, err := api.games.VoteForVideo(ctx, gameVideoId, tgId, request.Value)

// 	if err != nil {
// 		var sErr common.ServiceError
// 		if errors.As(err, &sErr) {
// 			if sErr.Code == common.ErrorForbidden {
// 				return c.JSON(http.StatusForbidden, m.ErrorResponse{
// 					Error: GameActionForbiddenMsg,
// 				})
// 			}
// 			if sErr.Code == common.ErrorDbData {
// 				return c.JSON(http.StatusBadRequest, m.ErrorResponse{
// 					Error: "invalid vote value",
// 				})
// 			}
// 		}

// 		c.Echo().Logger.Errorf("failed to vote for video: %w", err)
// 		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
// 			Error: SomethingWentWrongMsg,
// 		})
// 	}

// 	api.watchdog.EnqueueGame(game.ID)
// 	return c.JSON(http.StatusOK, vote)
// }

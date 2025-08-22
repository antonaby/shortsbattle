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

// func (api *GameApi) joinGame(c echo.Context) error {
// 	request := new(m.CreateGame)
// 	if err := c.Bind(request); err != nil {
// 		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
// 			Error: err.Error(),
// 		})
// 	}

// 	ctx := c.Request().Context()
// 	game, err := api.gm.JoinGame(ctx, request.ThemeId, request.PlayerId)
// 	if err != nil {
// 		c.Echo().Logger.Errorf("failed to create game: %v", err)

// 		var gsErr services.GameManagerError
// 		if errors.As(err, &gsErr) {
// 			if gsErr.Code == services.GMErrConstraintViolation {
// 				return c.JSON(http.StatusNotFound, m.ErrorResponse{
// 					Error: "Theme not found",
// 				})
// 			}
// 		}

// 		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
// 			Error: "Something went wrong",
// 		})
// 	}

// 	return c.JSON(http.StatusOK, game)
// }

// func (api *GameApi) getGame(c echo.Context) error {
// 	gameId, err := parseInt64(c.Param("id"))
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
// 			Error: err.Error(),
// 		})
// 	}

// 	ctx := c.Request().Context()
// 	details, err := api.gm.GetGame(ctx, gameId)
// 	if err != nil {
// 		c.Echo().Logger.Errorf("failed to get game: %v", err)

// 		var gsErr services.GameManagerError
// 		if errors.As(err, &gsErr) {
// 			if gsErr.Code == services.GMErrNotFound {
// 				return c.JSON(http.StatusNotFound, m.ErrorResponse{
// 					Error: "Game not found",
// 				})
// 			}
// 		}

// 		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
// 			Error: "Something went wrong",
// 		})
// 	}

// 	return c.JSON(http.StatusOK, details)
// }

// func (api *GameApi) submitVideo(c echo.Context) error {
// 	gameId, err := parseInt64(c.Param("gameId"))
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
// 			Error: err.Error(),
// 		})
// 	}

// 	request := new(m.SubmitVideoToGame)
// 	if err := c.Bind(request); err != nil {
// 		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
// 			Error: err.Error(),
// 		})
// 	}

// 	ctx := c.Request().Context()
// 	err = api.gm.SubmitVideo(ctx, db.CreateVideoParams{
// 		GameID:   gameId,
// 		PlayerID: request.PlayerId,
// 		VideoUrl: request.VideoUrl,
// 		IsActual: true,
// 	})

// 	if err != nil {
// 		c.Echo().Logger.Errorf("failed to submit video to game: %v", err)

// 		var gErr services.GameManagerError
// 		if errors.As(err, &gErr) {
// 			if gErr.Code == services.GMErrNotFound {
// 				return c.JSON(http.StatusNotFound, m.ErrorResponse{
// 					Error: "Game not found",
// 				})
// 			}
// 		}

// 		var rErr services.GameInstanceError
// 		if errors.As(err, &rErr) {
// 			if rErr.Code == services.GIErrWrongGameState {
// 				return c.JSON(http.StatusBadRequest, m.ErrorResponse{
// 					Error: "Lobby closed",
// 				})
// 			}

// 			if rErr.Code == services.GIErrConstraintViolation {
// 				return c.JSON(http.StatusNotFound, m.ErrorResponse{
// 					Error: "Game or Player not found",
// 				})
// 			}
// 		}

// 		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
// 			Error: "Something went wrong",
// 		})
// 	}

// 	return c.JSON(http.StatusOK, m.OkResponse{
// 		Msg: "Video has been submitted to the game",
// 	})
// }

// func (api *GameApi) submitVote(c echo.Context) error {
// 	gameId, err := parseInt64(c.Param("gameId"))
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
// 			Error: err.Error(),
// 		})
// 	}

// 	request := new(m.SubmitVoteToGame)
// 	if err := c.Bind(request); err != nil {
// 		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
// 			Error: err.Error(),
// 		})
// 	}

// 	ctx := c.Request().Context()
// 	err = api.gm.SubmitVote(ctx, db.CreateVoteParams{
// 		GameID:   gameId,
// 		VideoID:  request.VideoID,
// 		VoterID:  request.VoterID,
// 		Value:    request.Value,
// 		IsActual: true,
// 	})

// 	if err != nil {
// 		c.Echo().Logger.Errorf("failed to submit vote to game: %v", err)

// 		var gErr services.GameManagerError
// 		if errors.As(err, &gErr) {
// 			if gErr.Code == services.GMErrNotFound {
// 				return c.JSON(http.StatusNotFound, m.ErrorResponse{
// 					Error: "Game not found",
// 				})
// 			}
// 		}

// 		var rErr services.GameInstanceError
// 		if errors.As(err, &rErr) {
// 			if rErr.Code == services.GIErrWrongGameState {
// 				return c.JSON(http.StatusBadRequest, m.ErrorResponse{
// 					Error: "Voting closed",
// 				})
// 			}

// 			if rErr.Code == services.GIErrConstraintViolation {
// 				return c.JSON(http.StatusNotFound, m.ErrorResponse{
// 					Error: "Game, Player or Video not found",
// 				})
// 			}
// 		}

// 		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
// 			Error: "Something went wrong",
// 		})
// 	}

// 	return c.JSON(http.StatusOK, m.OkResponse{
// 		Msg: "Vote has been submitted to the game",
// 	})
// }

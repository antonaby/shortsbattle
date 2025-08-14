package api

import (
	"errors"
	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	m "github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/antonaby/shortsbattle/game-server/internal/services"

	"github.com/labstack/echo/v4"
)

type GameApi struct {
	gm *services.GameManager
}

func NewGameApi(gm *services.GameManager) *GameApi {
	return &GameApi{
		gm: gm,
	}
}

func (api *GameApi) Register(g *echo.Group) {
	v1group := g.Group("/v1")

	v1group.POST("/games", api.createGame)
	v1group.GET("/games/:id", api.getGame)
	v1group.PUT("/games/:gameId/players", api.addPlayerToGame)
	v1group.PUT("/games/:gameId/videos", api.submitVideo)
}

func (api *GameApi) createGame(c echo.Context) error {
	request := new(m.CreateGame)
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: err.Error(),
		})
	}

	ctx := c.Request().Context()
	game, err := api.gm.CreateGame(ctx, request.ThemeId)
	if err != nil {
		c.Echo().Logger.Errorf("failed to create game: %v", err)

		var gsErr services.GameManagerError
		if errors.As(err, &gsErr) {
			if gsErr.Code == services.GMErrConstraintViolation {
				return c.JSON(http.StatusNotFound, m.ErrorResponse{
					Error: "Theme not found",
				})
			}
		}

		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, game)
}

func (api *GameApi) getGame(c echo.Context) error {
	gameId, err := parseInt64(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: err.Error(),
		})
	}

	ctx := c.Request().Context()
	details, err := api.gm.GetGame(ctx, gameId)
	if err != nil {
		c.Echo().Logger.Errorf("failed to get game: %v", err)

		var gsErr services.GameManagerError
		if errors.As(err, &gsErr) {
			if gsErr.Code == services.GMErrNotFound {
				return c.JSON(http.StatusNotFound, m.ErrorResponse{
					Error: "Game not found",
				})
			}
		}

		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, details)
}

func (api *GameApi) addPlayerToGame(c echo.Context) error {
	gameId, err := parseInt64(c.Param("gameId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: err.Error(),
		})
	}

	request := new(m.AddPlayerToGame)
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: err.Error(),
		})
	}

	ctx := c.Request().Context()
	err = api.gm.AddPlayer(ctx, db.AddPlayerToGameParams{GameID: gameId, PlayerID: request.PlayerId})
	if err != nil {
		c.Echo().Logger.Errorf("failed to add player to game: %v", err)

		var gErr services.GameManagerError
		if errors.As(err, &gErr) {
			if gErr.Code == services.GMErrNotFound {
				return c.JSON(http.StatusNotFound, m.ErrorResponse{
					Error: "Game not found",
				})
			}
		}

		var rErr services.RoundError
		if errors.As(err, &rErr) {
			if rErr.Code == services.RoundErrTooManyPlayers || rErr.Code == services.RoundErrWrongGameState {
				return c.JSON(http.StatusBadRequest, m.ErrorResponse{
					Error: "Too many players or lobby closed",
				})
			}

			if rErr.Code == services.RoundErrConstraintViolation {
				return c.JSON(http.StatusNotFound, m.ErrorResponse{
					Error: "Game or Player not found",
				})
			}
		}

		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, m.OkResponse{
		Msg: "Player has been added to the game",
	})
}

func (api *GameApi) submitVideo(c echo.Context) error {
	gameId, err := parseInt64(c.Param("gameId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: err.Error(),
		})
	}

	request := new(m.SubmitVideoToGame)
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: err.Error(),
		})
	}

	ctx := c.Request().Context()
	err = api.gm.SubmitVideo(ctx, db.CreateVideoParams{
		GameID:   gameId,
		PlayerID: request.PlayerId,
		VideoUrl: request.VideoUrl,
	})

	if err != nil {
		c.Echo().Logger.Errorf("failed to submit video to game: %v", err)

		var gErr services.GameManagerError
		if errors.As(err, &gErr) {
			if gErr.Code == services.GMErrNotFound {
				return c.JSON(http.StatusNotFound, m.ErrorResponse{
					Error: "Game not found",
				})
			}
		}

		var rErr services.RoundError
		if errors.As(err, &rErr) {
			if rErr.Code == services.RoundErrWrongGameState {
				return c.JSON(http.StatusBadRequest, m.ErrorResponse{
					Error: "Lobby closed",
				})
			}

			if rErr.Code == services.RoundErrConstraintViolation {
				return c.JSON(http.StatusNotFound, m.ErrorResponse{
					Error: "Game or Player not found",
				})
			}
		}

		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, m.OkResponse{
		Msg: "Video has been submitted to the game",
	})
}

// func (api *GameApi) SubmitVote(c echo.Context) error {
// 	request := new(db.CreateVoteParams)
// 	if err := c.Bind(request); err != nil {
// 		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
// 			Error: err.Error(),
// 		})
// 	}

// 	ctx := c.Request().Context()
// 	vote, err := api.gm.SubmitVote(ctx, *request)
// 	if err != nil {
// 		c.Echo().Logger.Errorf("failed to submit vote: %v", err)
// 		return getErrorResponse(c, err)
// 	}

// 	return c.JSON(http.StatusOK, vote)
// }

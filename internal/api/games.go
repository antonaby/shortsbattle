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

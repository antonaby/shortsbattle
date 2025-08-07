package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/antonaby/shortsbattle/game-server/internal/services"
	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
    Error string `json:"error"`
}

type GameApi struct {
	gs *services.GameService
}

func NewGameApi(gs *services.GameService) *GameApi {
	return &GameApi{
		gs: gs,
	}
}

func (api *GameApi) Register(g *echo.Group) {
	v1group := g.Group("/v1")
	v1group.GET("/games", api.GetGames)
	v1group.GET("/player/:id", api.GetPlayer)
}

func (api *GameApi) GetGames(c echo.Context) error {
	ctx := c.Request().Context()
	return c.JSON(http.StatusOK, api.gs.GetGames(ctx))
}

func (api *GameApi) GetPlayer(c echo.Context) error {
	id := c.Param("id")
	playerId, err := strconv.ParseInt(id, 10, 32) 
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
	}

	ctx := c.Request().Context()
	player, err := api.gs.GetPlayer(ctx, int32(playerId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error: "Player not found",
			})
		}

		c.Echo().Logger.Errorf("failed to get player: %v", err)
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, player)
}

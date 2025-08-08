package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	m "github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/antonaby/shortsbattle/game-server/internal/services"
	"github.com/labstack/echo/v4"
)


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
	v1group.GET("/players/:id", api.GetPlayer)
	v1group.POST("/players", api.CreatePlayer)
}

func (api *GameApi) GetGames(c echo.Context) error {
	ctx := c.Request().Context()
	return c.JSON(http.StatusOK, api.gs.GetGames(ctx))
}

func (api *GameApi) CreatePlayer(c echo.Context) error {
	request := new(m.CreatePlayerRequest)
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: err.Error(),
		})
	}
	
	ctx := c.Request().Context()
	player, err := api.gs.CreatePlayer(ctx, *request)
	if err != nil {
		c.Echo().Logger.Errorf("failed to get player: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, player)
}

func (api *GameApi) GetPlayer(c echo.Context) error {
	id := c.Param("id")
	playerId, err := strconv.ParseInt(id, 10, 64) 
	if err != nil {
		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
			Error: err.Error(),
		})
	}

	ctx := c.Request().Context()
	player, err := api.gs.GetPlayer(ctx, int64(playerId))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, m.ErrorResponse{
				Error: "Player not found",
			})
		}

		c.Echo().Logger.Errorf("failed to get player: %v", err)
		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
			Error: "Something went wrong",
		})
	}

	return c.JSON(http.StatusOK, player)
}

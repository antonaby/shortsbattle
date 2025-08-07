package api

import (
	"net/http"

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
}

func (api *GameApi) GetGames(c echo.Context) error {
	return c.JSON(http.StatusOK, api.gs.GetGames())
}

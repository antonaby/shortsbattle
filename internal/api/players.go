package api

// import (
// 	"errors"
// 	"net/http"

// 	"github.com/antonaby/shortsbattle/game-server/internal/db"
// 	m "github.com/antonaby/shortsbattle/game-server/internal/models"
// 	"github.com/antonaby/shortsbattle/game-server/internal/services"
// 	"github.com/labstack/echo/v4"
// )

// type PlayersApi struct {
// 	ps *services.PlayersService
// }

// func NewPlayerApi(ps *services.PlayersService) *PlayersApi {
// 	return &PlayersApi{
// 		ps: ps,
// 	}
// }

// func (api *PlayersApi) Register(g *echo.Group) {
// 	v1group := g.Group("/v1")

// 	v1group.GET("/players/:id", api.getPlayer)
// 	v1group.POST("/players", api.createPlayer)
// }

// func (api *PlayersApi) createPlayer(c echo.Context) error {
// 	request := new(db.CreatePlayerParams)
// 	if err := c.Bind(request); err != nil {
// 		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
// 			Error: err.Error(),
// 		})
// 	}

// 	ctx := c.Request().Context()
// 	player, err := api.ps.CreatePlayer(ctx, *request)
// 	if err != nil {
// 		c.Echo().Logger.Errorf("failed to create player: %v", err)
// 		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
// 			Error: "Something went wrong",
// 		})
// 	}

// 	return c.JSON(http.StatusOK, player)
// }

// func (api *PlayersApi) getPlayer(c echo.Context) error {
// 	playerId, err := parseInt64(c.Param("id"))
// 	if err != nil {
// 		return c.JSON(http.StatusBadRequest, m.ErrorResponse{
// 			Error: err.Error(),
// 		})
// 	}

// 	ctx := c.Request().Context()
// 	player, err := api.ps.GetPlayer(ctx, playerId)
// 	if err != nil {
// 		c.Echo().Logger.Errorf("failed to get player: %v", err)

// 		var psErr services.PlayersServiceError
// 		if errors.As(err, &psErr) {
// 			if psErr.Code == services.PSErrNotFound {
// 				return c.JSON(http.StatusNotFound, m.ErrorResponse{
// 					Error: "Player not found",
// 				})
// 			}
// 		}

// 		return c.JSON(http.StatusInternalServerError, m.ErrorResponse{
// 			Error: "Something went wrong",
// 		})
// 	}

// 	return c.JSON(http.StatusOK, player)
// }

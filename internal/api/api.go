package api

import (
	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

func NewEchoServer() *echo.Echo {
	e := echo.New()
	e.Validator = NewCustomValidator()

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error != nil {
				log.Error().Err(v.Error).
					Str("URI", v.URI).
					Int("status", v.Status).
					Msg("request")
			} else {
				log.Debug().
					Str("URI", v.URI).
					Int("status", v.Status).
					Msg("request")
			}

			return nil
		},
	}))

	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	return e
}

type HttpApi struct {
	gm *services.GameManager
	ts *services.ThemeService
	vs *services.VideosService
}

func NewHttpApi(gm *services.GameManager, ts *services.ThemeService, vs *services.VideosService, g *echo.Group) *HttpApi {
	api := &HttpApi{
		gm: gm,
		ts: ts,
		vs: vs,
	}

	api.register(g)

	return api
}

func (api *HttpApi) register(g *echo.Group) {
	v1group := g.Group("/v1")

	v1group.PUT("/games/join", api.joinGame)
	v1group.PUT("/games/:id/submit", api.submitVideo)
	v1group.GET("/games/:id/videos", api.getVideosForGame) // TODO: return DTOs intead of DB Models
	v1group.PUT("/games/:id/vote", api.voteForVideo)       // TODO: return DTOs intead of DB Models

	v1group.POST("/themes", api.createTheme)
	v1group.GET("/themes", api.listAllThemes)
	v1group.GET("/themes/:id", api.getTheme)

	v1group.POST("/themes/:id/videos", api.createVideoRequest)

	v1group.POST("/videos", api.addVideo)
	v1group.GET("/videos/:id", api.getVideo)

	v1group.GET("/players/:id/videos", api.getVideosForPlayer)
}

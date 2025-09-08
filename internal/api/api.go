package api

import (
	"net/http"
	"strings"

	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/antonaby/shortsbattle/game-server/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

var JWTContextKey = "jwt"

type HttpApi struct {
	auth     *services.AuthService
	watchdog *services.GameWatchdog
	games    *services.GameManager
	themes   *services.ThemeService
	videos   *services.VideoService
}

func NewHttpApi(
	auth *services.AuthService,
	gwd *services.GameWatchdog,
	gm *services.GameManager,
	ts *services.ThemeService,
	vs *services.VideoService,
) *HttpApi {
	return &HttpApi{
		auth:     auth,
		watchdog: gwd,
		games:    gm,
		themes:   ts,
		videos:   vs,
	}
}

func (api *HttpApi) NewEchoServer() *echo.Echo {
	e := echo.New()
	e.Validator = NewCustomValidator()

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error != nil {
				log.Error().Err(v.Error).
					Stack().
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

	unsecuredApiGroup := e.Group("/api")
	api.addUnsecuredEndpoints(unsecuredApiGroup)

	securedApiGroup := e.Group("/api", api.authMiddleware)
	api.addSecuredEndpoints(securedApiGroup)

	return e
}

func (api *HttpApi) addUnsecuredEndpoints(g *echo.Group) {
	v1group := g.Group("/v1")

	v1group.POST("/auth", api.getTokenForTgUser)
}

func (api *HttpApi) addSecuredEndpoints(g *echo.Group) {
	v1group := g.Group("/v1")

	v1group.PUT("/games/join", api.joinGame)
	v1group.PUT("/games/:id/submit", api.submitVideo)
	v1group.GET("/games/:id/videos", api.getVideosForGame) // TODO: return DTOs intead of DB Models

	v1group.PUT("/votes/:id", api.voteForVideo) // TODO: return DTOs intead of DB Models

	// TODO: add player id as the owner of the theme
	v1group.POST("/themes", api.createTheme)
	v1group.GET("/themes", api.listAllThemes)
	v1group.GET("/themes/:id", api.getTheme)
	v1group.POST("/themes/:id/rounds", api.createRound)

	v1group.POST("/videos", api.addVideo)
	v1group.GET("/videos/:id", api.getVideo)

	v1group.GET("/me/videos", api.getVideosForPlayer)
}

func (api *HttpApi) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		h := c.Request().Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			return c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "missing authorization token",
			})
		}
		tokenStr := strings.TrimPrefix(h, "Bearer")
		tokenStr = strings.TrimSpace(tokenStr)

		token, err := api.auth.ParseAndValidateJwt([]byte(tokenStr))
		if err != nil {
			return c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid authorization token",
			})
		}

		c.Set(JWTContextKey, token)
		return next(c)
	}
}

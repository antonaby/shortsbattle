package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/antonaby/shortsbattle/game-server/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

type Watchdog interface {
	AdvanceGameNow(context.Context, int64) error
}

var JWTContextKey = "jwt"

type HttpApi struct {
	watchdog Watchdog
	auth     *services.AuthService
	games    *services.GameManager
	themes   *services.ThemeService
	videos   *services.VideoService
}

func NewHttpApi(
	watchdog Watchdog,
	auth *services.AuthService,
	gm *services.GameManager,
	ts *services.ThemeService,
	vs *services.VideoService,
) *HttpApi {
	return &HttpApi{
		watchdog: watchdog,
		auth:     auth,
		games:    gm,
		themes:   ts,
		videos:   vs,
	}
}

func (api *HttpApi) NewEchoServer() *echo.Echo {
	e := echo.New()
	e.Validator = NewCustomValidator()

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogError:     true,
		LogURI:       true,
		LogStatus:    true,
		LogMethod:    true,
		LogHost:      true,
		LogUserAgent: true,
		LogLatency:   true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error != nil {
				log.
					Error().
					Timestamp().
					Err(v.Error).
					Stack().
					Str("URI", v.URI).
					Int("status", v.Status).
					Str("method", v.Method).
					Str("host", v.Host).
					Str("user-agent", v.UserAgent).
					Dur("latency", v.Latency).
					Send()
			} else {
				log.Debug().
					Timestamp().
					Str("URI", v.URI).
					Int("status", v.Status).
					Str("method", v.Method).
					Str("host", v.Host).
					Str("user-agent", v.UserAgent).
					Dur("latency", v.Latency).
					Send()
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

	// TODO: secure
	adminApiGroup := e.Group("/internal")
	api.addInternalEndpoints(adminApiGroup)

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

// TODO: add an endpoint to mark unavailable video and exclude those video from the competition
func (api *HttpApi) addSecuredEndpoints(g *echo.Group) {
	v1group := g.Group("/v1")

	// game actions
	v1group.PUT("/games/join", api.joinGame)
	v1group.PUT("/games/:id/submit", api.submitVideo)
	v1group.PUT("/games/:id/mode", api.updateGameMode)
	v1group.GET("/games/:id/videos", api.getVideosForGame)
	v1group.GET("/games/:id/result", api.getGameResult)

	// themes
	v1group.GET("/themes", api.listAllThemes)
	v1group.GET("/themes/:id", api.getTheme)

	// votes
	v1group.PUT("/votes/:id", api.voteForVideo)

	// players and videos
	v1group.POST("/videos", api.addVideo)
	v1group.GET("/me/videos", api.getVideosForPlayer)
}

func (api *HttpApi) addInternalEndpoints(g *echo.Group) {
	v1group := g.Group("/v1")

	// themes
	v1group.POST("/themes", api.createTheme)
	v1group.POST("/themes/:id/rounds", api.createRound)
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

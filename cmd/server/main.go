package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/api"
	"github.com/antonaby/shortsbattle/game-server/internal/bot"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/services"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/antonaby/shortsbattle/game-server/internal/ws"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func configureLogger() {
	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.
		New(os.Stdout).
		With().
		Timestamp().
		Str("service", "shortsbattle").
		Str("env", os.Getenv("ENV")).
		Logger()
}

func configureEcho() *echo.Echo {
	e := echo.New()
	e.Validator = api.NewCustomValidator()

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

func main() {
	configureLogger()

	err := godotenv.Load()
	if err != nil {
		log.Error().Err(err).Msg(".env file not loaded")
	}

	dbManager, err := db.NewDbManager(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("Can't connect to Postgres")
	}
	defer dbManager.Close()

	redisClient, err := db.NewRedisClient()
	if err != nil {
		log.Fatal().Err(err).Msg("Can't connect to Redis")
	}
	defer redisClient.Close()

	ts := services.NewThemeService(dbManager)
	vs := services.NewVideosService(dbManager)
	ps := services.NewPlayersService(dbManager, redisClient)
	gm := services.NewGameManager(redisClient)

	cf, err := ws.NewCentrifugeServer()
	if err != nil {
		log.Fatal().Err(err).Msg("Can't create Centriguge server")
	}

	err = cf.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("Can't run Centriguge server")
	}

	botManager, err := bot.NewTgBotManager(ps, vs)
	if err != nil {
		log.Fatal().Err(err).Msg("Can't start TG Bot")
	}

	e := configureEcho()

	e.GET("/api/v1/games/updates", echo.WrapHandler(cf.Handler()))
	apiGroup := e.Group("/api")
	_ = api.NewThemesApi(ts, apiGroup)
	_ = api.NewVideosApi(vs, apiGroup)
	_ = api.NewGamesApi(gm, apiGroup)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		botManager.Start(ctx)
	}()

	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

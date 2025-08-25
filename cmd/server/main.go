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

	gameConfig := services.GameConfig{
		MaxPlayers:        5,
		LobbyState:        30 * time.Second,
		LobbyClosedBefore: 5 * time.Second,
		SubmittingState:   30 * time.Second,
		WatchingState:     30 * time.Second,
	}

	themeService := services.NewThemeService(dbManager)
	videoService := services.NewVideosService(dbManager)
	playerService := services.NewPlayersService(dbManager, redisClient)
	gameManager := services.NewGameManager(dbManager, gameConfig)

	centrifugeServer, err := ws.NewCentrifugeServer(gameManager)
	if err != nil {
		log.Fatal().Err(err).Msg("Can't create Centriguge server")
	}

	err = centrifugeServer.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("Can't run Centriguge server")
	}

	botManager, err := bot.NewTgBotManager(playerService, videoService)
	if err != nil {
		log.Fatal().Err(err).Msg("Can't start TG Bot")
	}

	gameStream := "games"
	gameConsumerGroup := "games-workers"
	err = db.EnsureStreamGroup(redisClient, gameStream, gameConsumerGroup)
	if err != nil {
		log.Fatal().Err(err).Msg("Can't create Redis Stream")
	}

	gameWatchdog := services.NewGameWatchdog(dbManager, redisClient, 1*time.Second, 100, gameStream, 1000)
	gameStateListener := services.NewGameStateListener(redisClient, gameManager, centrifugeServer, gameStream, gameConsumerGroup, "1", 10, 5*time.Second, 20*time.Second, 1000)

	e := configureEcho()

	e.GET("/api/v1/games/updates", echo.WrapHandler(centrifugeServer.Handler()))
	apiGroup := e.Group("/api")
	_ = api.NewThemesApi(themeService, apiGroup)
	_ = api.NewVideosApi(videoService, apiGroup)
	_ = api.NewGamesApi(gameManager, apiGroup)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err = gameStateListener.ReclaimPending(ctx)
	if err != nil {
		log.Fatal().Err(err).Msg("Can't reclaim pending messages")
	}

	go gameStateListener.Listen(ctx)
	go gameWatchdog.Run(ctx)
	go botManager.Start(ctx)

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

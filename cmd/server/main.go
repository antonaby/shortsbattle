package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/api"
	"github.com/antonaby/shortsbattle/game-server/internal/bot"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/services"
	"github.com/antonaby/shortsbattle/game-server/internal/watchdog"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"

	"github.com/antonaby/shortsbattle/game-server/internal/ws"
	"github.com/labstack/echo/v4"
)

func main() {
	configureLogger()

	err := godotenv.Load()
	if err != nil {
		log.Error().Err(err).Msg(".env file not loaded")
	}

	dbHost, err := db.GetDbDSN()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	dbManager, err := db.NewDbManager(dbHost)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	defer dbManager.Close()

	redisHost, err := db.GetRedisDSN()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	redisClient, err := db.NewRedisClient(redisHost)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	defer redisClient.Close()

	gameConfig := services.GameConfig{
		MaxPlayers:             2,
		MaxLobbyStage:          60 * time.Second,
		MaxLobbyFullStage:      3 * time.Second,
		MaxSubmitStage:         60 * time.Second,
		MaxSubmitCompleteStage: 3 * time.Second,
		MaxWatchStage:          600 * time.Second,
		MaxWatchCompleteStage:  3 * time.Second,
	}

	themeService := services.NewThemeService(dbManager)
	videoService := services.NewVideosService(dbManager)
	playerService := services.NewCachedPlayerService(dbManager, redisClient)
	gameManager := services.NewGameManager(dbManager, gameConfig)

	authConfig := services.AuthConfig{
		TokenExpTime: 300 * time.Minute,
		Issuer:       "shortsbattle",
		Audience:     "tg-mini-app",
	}

	keyManager, err := services.NewKeyManager()
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	authService := services.NewAuthService(keyManager, playerService, authConfig)

	wsConfig := ws.WsConnectionConfig{
		ConnectionExpTime: 1 * time.Minute,
	}

	centrifugeServer, err := ws.NewCentrifugeServer(gameManager, authService, wsConfig)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	err = centrifugeServer.Run()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	defer centrifugeServer.Shutdown()

	botManager, err := bot.NewTgBotManager(playerService, videoService)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	watchdogClient := watchdog.NewWatchdogClient(redisHost)
	defer watchdogClient.Close()

	advanceAtProcessor := watchdog.NewAdvanceGameAtProcessor(gameManager, centrifugeServer)
	advanceNowProcessor := watchdog.NewAdvanceGameNowProcessor(gameManager, centrifugeServer)
	watchdogWorker := watchdog.NewWatchdogWorker(redisHost, advanceAtProcessor, advanceNowProcessor)
	dbWatchdog, err := watchdog.NewDBWatchdog(dbManager, watchdogClient, dbHost, "status_updates")
	if err != nil {
		log.Fatal().Err(err).Send()
	}

	httpApi := api.NewHttpApi(watchdogClient, authService, gameManager, themeService, videoService)
	e := httpApi.NewEchoServer()
	e.GET("/api/v1/games/updates", echo.WrapHandler(centrifugeServer.Handler()))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	err = watchdogWorker.Run()
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	defer watchdogWorker.Shutdown()

	go botManager.Start(ctx)

	go dbWatchdog.Listen(ctx)
	err = dbWatchdog.ReprocessMissed(ctx)
	if err != nil {
		log.Fatal().Err(err).Send()
	}

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

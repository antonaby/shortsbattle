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

	dbManager, err := db.NewDbManager(nil)
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
		MaxPlayers:      2,
		MaxLobbyStage:   60 * time.Second,
		SubmittingState: 60 * time.Second,
		WatchingState:   600 * time.Second,
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
		log.Fatal().Err(err).Msg("Can't load JWK")
	}

	authService := services.NewAuthService(keyManager, playerService, authConfig)

	wsConfig := ws.WsConnectionConfig{
		ConnectionExpTime: 1 * time.Minute,
	}

	centrifugeServer, err := ws.NewCentrifugeServer(gameManager, authService, wsConfig)
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

	redisHost, err := db.GetRedisDSN()
	if err != nil {
		log.Fatal().Err(err).Msg("can't get redis host")
	}

	advanceProcessor := watchdog.NewAdvanceGameProcessor(gameManager, centrifugeServer)
	watchdogWorker := watchdog.NewWatchdogWorker(redisHost, advanceProcessor)
	watchdogPublisher := watchdog.NewWatchdogPublisher(redisHost)
	defer watchdogPublisher.Close()

	httpApi := api.NewHttpApi(watchdogPublisher, authService, gameManager, themeService, videoService)
	e := httpApi.NewEchoServer()
	e.GET("/api/v1/games/updates", echo.WrapHandler(centrifugeServer.Handler()))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go watchdogWorker.Run()
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

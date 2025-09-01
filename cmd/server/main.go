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

	authConfig := services.AuthConfig{
		TokenExpTime: 5 * time.Minute,
		Issuer:       "shortsbattle",
		Audience:     "tg-mini-app",
	}

	keyManager, err := services.NewKeyManager()
	if err != nil {
		log.Fatal().Err(err).Msg("Can't load JWK")
	}

	authService := services.NewAuthService(keyManager, authConfig)

	gameConfig := services.GameConfig{
		MaxPlayers:        5,
		LobbyState:        10 * time.Second,
		LobbyClosedBefore: 5 * time.Second,
		SubmittingState:   10 * time.Second,
		WatchingState:     20 * time.Second,
	}

	themeService := services.NewThemeService(dbManager)
	videoService := services.NewVideosService(dbManager)
	playerService := services.NewPlayersService(dbManager, redisClient)
	gameManager := services.NewGameManager(dbManager, redisClient, videoService, gameConfig)

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

	httpApi := api.NewHttpApi(authService, gameManager, themeService, videoService)
	e := httpApi.NewEchoServer()
	e.GET("/api/v1/games/updates", echo.WrapHandler(centrifugeServer.Handler()))
	
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

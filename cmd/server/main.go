package main

import (
	"context"
	l "log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/api"
	"github.com/antonaby/shortsbattle/game-server/internal/bot"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/services"
	"github.com/joho/godotenv"

	"github.com/antonaby/shortsbattle/game-server/internal/ws"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		l.Println("No .env file found")
	}

	dbManager, err := db.NewDbManager(context.Background())
	if err != nil {
		log.Fatal("Can't connect to DB")
	}
	defer dbManager.Close()

	redisClient, err := db.NewRedisClient()
	if err != nil{
		log.Fatal("Can't connect to Redis")
	}
	defer redisClient.Close()

	ts := services.NewThemeService(dbManager)
	vs := services.NewVideosService(dbManager)
	ps := services.NewPlayersService(dbManager, redisClient)

	// gm := services.NewGameManager(dbManager)

	cf, err := ws.NewCentrifugeServer()
	if err != nil {
		log.Fatal("Can't create Centriguge router")
	}
	err = cf.Run()
	if err != nil {
		log.Fatal("Can't start Centriguge router")
	}

	botManager, err := bot.NewTgBotManager(ps, vs)
	if err != nil {
		log.Fatal("Can't start Tg Bot")
	}

	//gm.SetEventPublisher(cf)

	e := echo.New()
	e.Logger.SetLevel(log.INFO)
	e.Validator = api.NewCustomValidator()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	e.GET("/api/v1/games/updates", echo.WrapHandler(cf.Handler()))

	apiGroup := e.Group("/api")
	_ = api.NewThemesApi(ts, apiGroup)
	_ = api.NewVideosApi(vs, apiGroup)

	// gameApi := api.NewGameApi(gm)
	// gameApi.Register(apiGroup)
	// playerApi := api.NewPlayerApi(ps)
	// playerApi.Register(apiGroup)

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

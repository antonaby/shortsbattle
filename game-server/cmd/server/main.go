package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/api"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/services"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func main() {
	dbManager, err := db.NewDbManager(context.Background())
	if err != nil {
		log.Fatal("Can't connect to DB")
	}
	defer dbManager.Close()

	ts := services.NewThemeService(dbManager)
	gm := services.NewGameManager(dbManager)
	ps := services.NewPlayersService(dbManager)

	cs, err := api.NewCentrifugeServer()
	if err != nil {
		log.Fatal("Can't create Centriguge router")
	}

	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	e.GET("/api/v1/join", echo.WrapHandler(cs.Handler()))

	apiGroup := e.Group("/api")

	themesApi := api.NewThemesApi(ts)
	themesApi.Register(apiGroup)
	gameApi := api.NewGameApi(gm)
	gameApi.Register(apiGroup)
	playerApi := api.NewPlayerApi(ps)
	playerApi.Register(apiGroup)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

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

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	routes "github.com/antonaby/shortsbattle/game-server/internal/http"
	"github.com/antonaby/shortsbattle/game-server/internal/services"
)

func main() {
	dbManager, err := db.NewDbManager(context.Background())
	if err != nil {
		log.Fatal("Can't connect to DB")
	}

	cr, err := routes.NewCentrifugeRouter()
	if err != nil {
		log.Fatal("Can't create Centriguge router")
	}

	gs := services.NewGameService(dbManager)
	router := routes.NewHttpRouter(cr, gs)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router.Handler(),
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Server forced to shutdown: %v", err)
		}
	}()

	log.Printf("Server is ready to handle requests at %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on %s: %v", srv.Addr, err)
	}

	log.Println("Server stopped")
}

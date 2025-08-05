package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/centrifugal/centrifuge"
)

type Category struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

var categories = []Category{
    {
        Name:        "The most cute cat 🐈",
        Description: "A game about the cutest cat in the world. 🐱",
    },
    {
        Name:        "Funniest fail video 😂",
        Description: "Submit a hilarious fail that makes everyone laugh!",
    },
    {
        Name:        "Best dance move 💃",
        Description: "Show off your craziest or smoothest dance step.",
    },
    {
        Name:        "Unexpected twist 🎭",
        Description: "Videos that take a surprising turn. Shock us!",
    },
    {
        Name:        "Cutest baby animal 🐾",
        Description: "Puppies, kittens, ducklings... bring the awws!",
    },
    {
        Name:        "Most epic moment ⚡",
        Description: "Highlight something legendary, heroic, or just cool.",
    },
    {
        Name:        "Mind-blowing magic trick 🎩✨",
        Description: "Is it real? Is it edited? Blow our minds!",
    },
    {
        Name:        "Satisfying video 🍰",
        Description: "Soap cutting, symmetry, pouring — we want chill.",
    },
    {
        Name:        "Cringe overload 😬",
        Description: "Bring the secondhand embarrassment in a fun way.",
    },
    {
        Name:        "Best pet reaction 🐶😲",
        Description: "Pets doing something wild, unexpected, or smart!",
    },
}

func authMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cred := &centrifuge.Credentials{
			UserID: "",
		}
		newCtx := centrifuge.SetCredentials(ctx, cred)
		r = r.WithContext(newCtx)
		h.ServeHTTP(w, r)
	})
}

func newRouter(centrifugeHandler *centrifuge.WebsocketHandler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/games", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(categories)
	})

	mux.Handle("/v1/join", authMiddleware(centrifugeHandler))

	return mux
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	node, err := centrifuge.New(centrifuge.Config{})
	if err != nil {
		log.Fatal(err)
	}

	node.OnConnect(func(client *centrifuge.Client) {
		transportName := client.Transport().Name()
		transportProto := client.Transport().Protocol()
		log.Printf("client connected via %s (%s)", transportName, transportProto)

		client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
			log.Printf("client subscribes on channel %s", e.Channel)
			cb(centrifuge.SubscribeReply{}, nil)
		})

		client.OnPublish(func(e centrifuge.PublishEvent, cb centrifuge.PublishCallback) {
			log.Printf("client publishes into channel %s: %s", e.Channel, string(e.Data))
			cb(centrifuge.PublishReply{}, nil)
		})

		client.OnDisconnect(func(e centrifuge.DisconnectEvent) {
			log.Printf("client disconnected")
		})
	})

	if err := node.Run(); err != nil {
		log.Fatal(err)
	}

	wsHandler := centrifuge.NewWebsocketHandler(node, centrifuge.WebsocketConfig{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for simplicity, adjust as needed
		},
	})
	router := newRouter(wsHandler)

	corsRouter := corsMiddleware(router)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: corsRouter,
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


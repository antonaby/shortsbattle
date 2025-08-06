package http

import (
	"encoding/json"
	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/services"
)

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

type HttpRouter struct {
	mux              *http.ServeMux
	centrifugeRouter *CentrifugeRouter
	gameService      *services.GameService
}

func NewHttpRouter(centrifugeRouter *CentrifugeRouter, gameService *services.GameService) *HttpRouter {
	mux := http.NewServeMux()
	r := &HttpRouter{
		mux:              mux,
		centrifugeRouter: centrifugeRouter,
		gameService:      gameService,
	}

	// Define routes
	mux.HandleFunc("/v1/games", r.handleGameList)
	mux.HandleFunc("/v1/health", r.handleHealthCheck)
	mux.Handle("/v1/join", centrifugeRouter.Handler())

	return r
}

func (rt HttpRouter) Handler() http.Handler {
	mux := rt.mux
	handler := corsMiddleware(mux)
	return handler
}

func (rt HttpRouter) handleGameList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rt.gameService.GetGames())
}

func (rt HttpRouter) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "ok"}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

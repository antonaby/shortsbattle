package http

import (
	"encoding/json"
	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/services"
)

type HttpRouter struct {
	Mux         *http.ServeMux
	gameService *services.GameService
}

func NewHttpRouter(gameService *services.GameService) *HttpRouter {
	mux := http.NewServeMux()
	r := &HttpRouter{
		Mux:         mux,
		gameService: gameService,
	}

	// Define routes
	mux.HandleFunc("/v1/games", r.handleGameList)
	mux.HandleFunc("/v1/health", r.handleHealthCheck)

	return r
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

package ws

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/antonaby/shortsbattle/game-server/internal/services"
	"github.com/centrifugal/centrifuge"
)

type JoinGameRequest struct {
	UserID int `json:"user_id"`
	GameID int `json:"game_id"`
}

type CentrifugeServer struct {
	Node *centrifuge.Node
	gm   *services.GameManager
}

func NewCentrifugeServer(gm *services.GameManager) (*CentrifugeServer, error) {
	node, err := centrifuge.New(centrifuge.Config{})
	if err != nil {
		return nil, err
	}

	r := &CentrifugeServer{
		Node: node,
		gm:   gm,
	}

	node.OnConnecting(r.handleConnecting)
	node.OnConnect(r.handleConnection)

	return r, nil
}

func (r CentrifugeServer) Handler() http.Handler {
	wsHandler := centrifuge.NewWebsocketHandler(r.Node, centrifuge.WebsocketConfig{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	})

	return wsHandler
}

func (r *CentrifugeServer) Run() error {
	if err := r.Node.Run(); err != nil {
		return err
	}

	return nil
}

func (r *CentrifugeServer) PublishGameUpdate(channel string, upd models.GameUpdate) error {
	jsonBytes, err := json.Marshal(upd)
	if err != nil {
		return err
	}

	_, err = r.Node.Publish(channel, jsonBytes)
	return err
}

func (r *CentrifugeServer) handleConnecting(ctx context.Context, e centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
	return centrifuge.ConnectReply{
		ClientSideRefresh: true,
		Credentials: &centrifuge.Credentials{
			UserID: "test_1",
		},
	}, nil
}

func (r *CentrifugeServer) handleConnection(client *centrifuge.Client) {
	client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
		var req JoinGameRequest
		err := json.Unmarshal(e.Data, &req)
		if err != nil {
			cb(centrifuge.SubscribeReply{}, centrifuge.ErrorBadRequest)
			return
		}

		cb(centrifuge.SubscribeReply{}, nil)
	})

	client.OnUnsubscribe(func(e centrifuge.UnsubscribeEvent) {
		idStr := strings.TrimPrefix(e.Channel, "game_")
		gameId, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			panic(err)
		}

		_ = r.gm.PlayerLeave(context.Background(), gameId, 1)
	})

	client.OnPublish(func(e centrifuge.PublishEvent, cb centrifuge.PublishCallback) {
		log.Printf("client publishes into channel %s: %s", e.Channel, string(e.Data))
		cb(centrifuge.PublishReply{}, nil)
	})

	client.OnDisconnect(func(e centrifuge.DisconnectEvent) {

	})
}

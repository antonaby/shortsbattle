package ws

import (
	"context"
	"encoding/json"

	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/centrifugal/centrifuge"
)

type JoinGameRequest struct {
	UserID int `json:"user_id"`
	GameID int `json:"game_id"`
}

type CentrifugeServer struct {
	Node *centrifuge.Node
}

func NewCentrifugeServer() (*CentrifugeServer, error) {
	node, err := centrifuge.New(centrifuge.Config{})
	if err != nil {
		return nil, err
	}

	r := &CentrifugeServer{
		Node: node,
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
		cb(centrifuge.SubscribeReply{}, nil)
	})

	client.OnUnsubscribe(func(e centrifuge.UnsubscribeEvent) {
		
	})

	client.OnDisconnect(func(e centrifuge.DisconnectEvent) {

	})
}

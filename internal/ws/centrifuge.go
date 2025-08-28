package ws

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/antonaby/shortsbattle/game-server/internal/services"
	"github.com/centrifugal/centrifuge"

	"github.com/rs/zerolog/log"
)

type CentrifugeServer struct {
	node    *centrifuge.Node
	manager *services.GameManager
}

func NewCentrifugeServer(manager *services.GameManager) (*CentrifugeServer, error) {
	node, err := centrifuge.New(centrifuge.Config{})
	if err != nil {
		return nil, err
	}

	r := &CentrifugeServer{
		node:    node,
		manager: manager,
	}

	node.OnConnecting(r.handleConnecting)
	node.OnConnect(r.handleConnection)

	return r, nil
}

func (r CentrifugeServer) Handler() http.Handler {
	wsHandler := centrifuge.NewWebsocketHandler(r.node, centrifuge.WebsocketConfig{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	})

	return wsHandler
}

func (cf *CentrifugeServer) Run() error {
	if err := cf.node.Run(); err != nil {
		return err
	}

	return nil
}

func (cf *CentrifugeServer) PublishGameUpdate(channel string, upd models.GameUpdate) error {
	jsonBytes, err := json.Marshal(upd)
	if err != nil {
		return err
	}

	_, err = cf.node.Publish(channel, jsonBytes)
	return err
}

func (cf *CentrifugeServer) handleConnecting(ctx context.Context, e centrifuge.ConnectEvent) (centrifuge.ConnectReply, error) {
	return centrifuge.ConnectReply{
		ClientSideRefresh: true,
		Credentials: &centrifuge.Credentials{
			UserID: "test_1",
		},
	}, nil
}

func (cf *CentrifugeServer) handleConnection(client *centrifuge.Client) {
	client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		gameId, err := services.ParseCfChannelName(e.Channel)
		if err != nil {
			log.Error().Err(err).Msgf("can't parse channel name: %s", e.Channel)
			cb(centrifuge.SubscribeReply{}, centrifuge.ErrorBadRequest)
			return
		}

		// TODO: add proper player id
		details, err := cf.manager.GetGameDetailsForPlayer(ctx, gameId, 1)
		if err != nil {
			log.Error().Err(err).Msgf("can't get game for channel: %s", e.Channel)

			var gErr common.ServiceError
			if errors.As(err, &gErr) {
				if gErr.Code == common.ErrorDbNotFound {
					cb(centrifuge.SubscribeReply{}, centrifuge.ErrorPermissionDenied)
					return
				}
			}

			cb(centrifuge.SubscribeReply{}, centrifuge.ErrorInternal)
			return
		}

		updBytes, err := json.Marshal(details)
		if err != nil {
			log.Error().Err(err).Msgf("failed to convert game update to bytes: %s", e.Channel)
			cb(centrifuge.SubscribeReply{}, centrifuge.ErrorInternal)
			return
		}

		cb(centrifuge.SubscribeReply{
			Options: centrifuge.SubscribeOptions{
				Data: updBytes,
			},
		}, nil)
	})

	client.OnUnsubscribe(func(e centrifuge.UnsubscribeEvent) {

	})

	client.OnDisconnect(func(e centrifuge.DisconnectEvent) {

	})
}

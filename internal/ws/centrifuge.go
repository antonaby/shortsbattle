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

type WsConnectionConfig struct {
	ConnectionExpTime time.Duration
}

type CentrifugeServer struct {
	node    *centrifuge.Node
	manager *services.GameManager
	auth    *services.AuthService
	config  WsConnectionConfig
}

func NewCentrifugeServer(manager *services.GameManager, auth *services.AuthService, config WsConnectionConfig) (*CentrifugeServer, error) {
	node, err := centrifuge.New(centrifuge.Config{})
	if err != nil {
		return nil, err
	}

	r := &CentrifugeServer{
		node:    node,
		manager: manager,
		auth:    auth,
		config:  config,
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
	token, err := cf.auth.ParseAndValidateJwt([]byte(e.Token))
	if err != nil {
		return centrifuge.ConnectReply{}, centrifuge.ErrorUnauthorized
	}

	sub, ok := token.Subject()
	if !ok {
		return centrifuge.ConnectReply{}, centrifuge.ErrorUnauthorized
	}

	return centrifuge.ConnectReply{
		ClientSideRefresh: true,
		Credentials: &centrifuge.Credentials{
			UserID:   sub,
			ExpireAt: time.Now().Add(cf.config.ConnectionExpTime).Unix(),
		},
	}, nil
}

func (cf *CentrifugeServer) handleConnection(client *centrifuge.Client) {
	// TODO: add online/ofline player statuses
	// TODO: add private channel
	client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		gameId, err := services.ParseCfChannelName(e.Channel)
		if err != nil {
			log.Error().Err(err).Stack().Msgf("can't parse channel name: %s", e.Channel)
			cb(centrifuge.SubscribeReply{}, centrifuge.ErrorBadRequest)
			return
		}

		tgId, err := common.ParseTgId(client.UserID())
		if err != nil {
			log.Error().Err(err).Stack().Msgf("can't parse user id: %s", client.UserID())
			cb(centrifuge.SubscribeReply{}, centrifuge.ErrorBadRequest)
			return
		}

		details, err := cf.manager.GetGameDetailsForPlayer(ctx, gameId, tgId)
		if err != nil {
			var gErr common.ServiceError
			if errors.As(err, &gErr) {
				if gErr.Code == common.ErrorForbidden {
					cb(centrifuge.SubscribeReply{}, centrifuge.ErrorPermissionDenied)
					return
				}
			}

			log.Error().Err(err).Stack().Msgf("can't get game for channel: %s", e.Channel)
			cb(centrifuge.SubscribeReply{}, centrifuge.ErrorInternal)
			return
		}

		updBytes, err := json.Marshal(details)
		if err != nil {
			log.Error().Err(err).Stack().Msgf("failed to convert game update to bytes: %s", e.Channel)
			cb(centrifuge.SubscribeReply{}, centrifuge.ErrorInternal)
			return
		}

		cb(centrifuge.SubscribeReply{
			Options: centrifuge.SubscribeOptions{
				Data: updBytes,
			},
		}, nil)
	})

	client.OnRefresh(func(e centrifuge.RefreshEvent, cb centrifuge.RefreshCallback) {
		_, err := cf.auth.ParseAndValidateJwt([]byte(e.Token))
		if err != nil {
			cb(centrifuge.RefreshReply{}, centrifuge.ErrorUnauthorized)
			return
		}

		cb(centrifuge.RefreshReply{
			Expired:  false,
			ExpireAt: time.Now().Add(cf.config.ConnectionExpTime).Unix(),
		}, nil)
	})

	client.OnUnsubscribe(func(e centrifuge.UnsubscribeEvent) {

	})

	client.OnDisconnect(func(e centrifuge.DisconnectEvent) {

	})
}

package ws

import (
	"context"
	"encoding/json"
	"fmt"

	"time"

	"net/http"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/antonaby/shortsbattle/game-server/internal/services"
	"github.com/centrifugal/centrifuge"
	"github.com/rs/zerolog/log"
)

const (
	ErrOnlineStatus = "failed to update players online status: %s"
	ErrChannelName  = "failed to parse channel name: %s"
)

func cfError(code common.ErrorCode, msg string, err error) error {
	return common.ServiceError{
		Code:    code,
		Message: fmt.Sprintf("centrifuge: %s", msg),
		Cause:   err,
	}
}

func logError(err error, msgf string, args ...any) {
	log.Error().Stack().Err(err).Msgf(msgf, args...)
}

type WsConnectionConfig struct {
	ClientPingInterval  time.Duration
	ConnectionExpTime   time.Duration
	SubscriptionExpTime time.Duration
}

type CentrifugeServer struct {
	node    *centrifuge.Node
	manager *services.GameManager
	auth    *services.AuthService
	players *services.CachedPlayerService
	config  WsConnectionConfig
}

func NewCentrifugeServer(manager *services.GameManager, auth *services.AuthService, players *services.CachedPlayerService, config WsConnectionConfig) (*CentrifugeServer, error) {
	node, err := centrifuge.New(centrifuge.Config{
		ClientPresenceUpdateInterval: config.ClientPingInterval,
	})
	if err != nil {
		return nil, cfError(common.ErrorInit, "failed to create server", err)
	}

	r := &CentrifugeServer{
		node:    node,
		manager: manager,
		auth:    auth,
		players: players,
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
		return cfError(common.ErrorInit, "failed to start server", err)
	}

	return nil
}

func (cf *CentrifugeServer) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := cf.node.Shutdown(ctx); err != nil {
		return cfError(common.ErrorInit, "failed to stop server", err)
	}

	return nil
}

func (cf *CentrifugeServer) PublishGameUpdate(upd models.GameUpdate) error {
	jsonBytes, err := json.Marshal(upd)
	if err != nil {
		return cfError(common.ErrorMarshal, "failed to marshal palyload", err)
	}

	channel := models.GetCfChannelName(upd.GameID)
	_, err = cf.node.Publish(channel, jsonBytes)
	if err != nil {
		return cfError(common.ErrorPublish, "failed to publish game update", err)
	}

	return nil
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

// TODO: add private channel
func (cf *CentrifugeServer) handleConnection(client *centrifuge.Client) {
	tgId, err := common.ParseTgId(client.UserID())
	if err != nil {
		logError(err, "failed to parse user id: %s", client.UserID())
		client.Disconnect(centrifuge.DisconnectServerError)
		return
	}

	err = cf.updatePlayerStatus(tgId, true)
	if err != nil {
		client.Disconnect(centrifuge.DisconnectServerError)
		return
	}

	client.OnSubscribe(func(e centrifuge.SubscribeEvent, cb centrifuge.SubscribeCallback) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		gameId, err := models.ParseCfChannelName(e.Channel)
		if err != nil {
			logError(err, ErrChannelName, e.Channel)
			cb(centrifuge.SubscribeReply{}, centrifuge.ErrorBadRequest)
			return
		}

		details, err := cf.manager.GetGameDetailsForPlayer(ctx, gameId, tgId)
		if err != nil {
			if sErr, ok := common.IsServErr(err); ok && sErr.Code == common.ErrorForbidden {
				cb(centrifuge.SubscribeReply{}, centrifuge.ErrorPermissionDenied)
				return
			}

			logError(err, "failed to get game state: %s", e.Channel)
			cb(centrifuge.SubscribeReply{}, centrifuge.ErrorInternal)
			return
		}

		err = cf.manager.SetPlayerOnlineStatus(ctx, gameId, tgId, true)
		if err != nil {
			if sErr, ok := common.IsServErr(err); ok && sErr.Code == common.ErrorNotFound {
				cb(centrifuge.SubscribeReply{}, centrifuge.ErrorPermissionDenied)
				return
			}

			logError(err, ErrOnlineStatus, client.UserID())
			cb(centrifuge.SubscribeReply{}, centrifuge.ErrorInternal)
			return
		}

		updBytes, err := json.Marshal(details)
		if err != nil {
			logError(err, "failed to convert game update to bytes: %s", e.Channel)
			cb(centrifuge.SubscribeReply{}, centrifuge.ErrorInternal)
			return
		}

		cb(centrifuge.SubscribeReply{
			ClientSideRefresh: true,
			Options: centrifuge.SubscribeOptions{
				ExpireAt: time.Now().Add(cf.config.SubscriptionExpTime).Unix(),
				Data:     updBytes,
			},
		}, nil)
	})

	client.OnSubRefresh(func(e centrifuge.SubRefreshEvent, cb centrifuge.SubRefreshCallback) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		gameId, err := models.ParseCfChannelName(e.Channel)
		if err != nil {
			logError(err, ErrChannelName, e.Channel)
			cb(centrifuge.SubRefreshReply{}, centrifuge.ErrorBadRequest)
			return
		}

		err = cf.manager.SetPlayerOnlineStatus(ctx, gameId, tgId, true)
		if err != nil {
			if sErr, ok := common.IsServErr(err); ok && sErr.Code == common.ErrorNotFound {
				cb(centrifuge.SubRefreshReply{}, centrifuge.ErrorPermissionDenied)
				return
			}

			logError(err, ErrOnlineStatus, client.UserID())
			cb(centrifuge.SubRefreshReply{}, centrifuge.ErrorInternal)
			return
		}

		cb(centrifuge.SubRefreshReply{
			ExpireAt: time.Now().Add(cf.config.SubscriptionExpTime).Unix(),
		}, nil)
	})

	client.OnUnsubscribe(func(e centrifuge.UnsubscribeEvent) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		gameId, err := models.ParseCfChannelName(e.Channel)
		if err != nil {
			logError(err, ErrChannelName, e.Channel)
			return
		}

		err = cf.manager.SetPlayerOnlineStatus(ctx, gameId, tgId, false)
		if err != nil {
			if sErr, ok := common.IsServErr(err); ok && sErr.Code == common.ErrorNotFound {
				return
			}

			logError(err, ErrOnlineStatus, client.UserID())
		}
	})

	client.OnRefresh(func(e centrifuge.RefreshEvent, cb centrifuge.RefreshCallback) {
		_, err := cf.auth.ParseAndValidateJwt([]byte(e.Token))
		if err != nil {
			cb(centrifuge.RefreshReply{}, centrifuge.ErrorUnauthorized)
			return
		}

		err = cf.updatePlayerStatus(tgId, true)
		if err != nil {
			client.Disconnect(centrifuge.DisconnectServerError)
			return
		}

		cb(centrifuge.RefreshReply{
			Expired:  false,
			ExpireAt: time.Now().Add(cf.config.ConnectionExpTime).Unix(),
		}, nil)
	})

	client.OnAlive(func() {
		cf.updatePlayerStatus(tgId, true)
	})

	client.OnDisconnect(func(e centrifuge.DisconnectEvent) {
		cf.updatePlayerStatus(tgId, false)
	})
}

// TODO: if false set all active games to false
func (cf *CentrifugeServer) updatePlayerStatus(tgId int64, status bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := cf.players.SetPlayerOnlineStatus(ctx, tgId, status)
	if err != nil {
		if sErr, ok := common.IsServErr(err); ok && sErr.Code == common.ErrorNotFound {
			return nil
		}

		logError(err, ErrOnlineStatus, fmt.Sprint(tgId))
		return err
	}

	return nil
}

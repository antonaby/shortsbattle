package ws

import (
	"log"
	"net/http"

	"github.com/centrifugal/centrifuge"
)

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

	node.OnConnect(r.handleConnection)

	return r, nil
}

func (r CentrifugeServer) Handler() http.Handler {
	wsHandler := centrifuge.NewWebsocketHandler(r.Node, centrifuge.WebsocketConfig{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for simplicity, adjust as needed
		},
	})

	return authMiddleware(wsHandler)
}

func (r *CentrifugeServer) Run() error {
	if err := r.Node.Run(); err != nil {
		return err
	}

	return nil
}

func (r *CentrifugeServer) Publish(channel string, data []byte) (centrifuge.PublishResult, error) {
	return r.Node.Publish(channel, data)
}

func (r *CentrifugeServer) handleConnection(client *centrifuge.Client) {
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
}

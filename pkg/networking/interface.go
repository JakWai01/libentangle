package networking

import (
	"sync"

	"github.com/alphahorizonio/libentangle/internal/handlers"
	"github.com/alphahorizonio/libentangle/pkg/signaling"
	"github.com/pion/webrtc/v3"
	"nhooyr.io/websocket"
)

var (
	manager *handlers.ClientManager
)

type NoConnectionEstablished struct{}

func (m *NoConnectionEstablished) Error() string {
	return "No connection established so far. Either the Connect() has not been called yet or the connection was still in the making"
}

func Connect(community string, f func(msg webrtc.DataChannelMessage), onConnected func()) {
	manager = handlers.NewClientManager(onConnected)

	client := signaling.NewSignalingClient(
		func(conn *websocket.Conn, uuid string) error {
			return manager.HandleAcceptance(conn, uuid)
		},
		func(conn *websocket.Conn, data []byte, uuid string, wg *sync.WaitGroup) error {
			return manager.HandleIntroduction(conn, data, uuid, wg, f)
		},
		func(conn *websocket.Conn, data []byte, candidates *chan string, wg *sync.WaitGroup, uuid string) error {
			return manager.HandleOffer(conn, data, candidates, wg, uuid, f)
		},
		func(data []byte, candidates *chan string, wg *sync.WaitGroup) error {
			return manager.HandleAnswer(data, candidates, wg)
		},
		func(data []byte, candidates *chan string) error {
			return manager.HandleCandidate(data, candidates)
		},
		func() error {
			return manager.HandleResignation()
		},
	)

	go func() {
		go client.HandleConn("localhost:9090", community, f)
	}()
}

func WriteToDataChannel(p []byte) (int, error) {
	if err := Write(p); err != nil {
		return 0, err
	}
	return len(p), nil
}

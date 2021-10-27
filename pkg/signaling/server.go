package signaling

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	api "github.com/alphahorizon/libentangle/pkg/api/websockets/v1"
	"nhooyr.io/websocket"
)

type SignalingServer struct {
	lock        sync.Mutex
	communities map[string][]string
	macs        map[string]bool
	connections map[string]websocket.Conn
}

func NewSignalingServer() *SignalingServer {
	return &SignalingServer{
		communities: map[string][]string{},
		macs:        map[string]bool{},
		connections: map[string]websocket.Conn{},
	}
}

func (s *SignalingServer) HandleConn(conn websocket.Conn) {

	go func() {
		for {
			// Read message from connection
			_, data, err := conn.Read(context.Background())
			if err != nil {
				log.Fatal(err)
			}

			// Parse message
			var v api.Message
			if err := json.Unmarshal(data, &v); err != nil {
				log.Fatal(err)
			}

			fmt.Println(v)

			// Handle different message types
			switch v.Opcode {
			// TODO
			}
		}
	}()
}

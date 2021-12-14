package networking

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/alphahorizon/libentangle/pkg/signaling"
	"github.com/pion/webrtc/v3"
	"nhooyr.io/websocket"
)

var (
	manager *signaling.ClientManager
)

type NoConnectionEstablished struct{}

func (m *NoConnectionEstablished) Error() string {
	return "No connection established so far. Either the Connect() has not been called yet or the connection was still in the making"
}

func Connect(community string, f func(msg webrtc.DataChannelMessage)) {
	manager = signaling.NewClientManager()

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

func ReadWriter() {
	f, err := os.Open("test.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buf := make([]byte, 10)
	for {
		n, err := f.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			break
		}
		if n > 0 {
			fmt.Println(string(buf[:n]))
			Write(buf[:n])
		}
	}
}

func Write(msg []byte) error {
	if manager != nil {
		manager.SendMessage(msg)
		return nil
	}
	return &NoConnectionEstablished{}
}

func WriteUnicast(msg []byte, mac string) error {
	if manager != nil {
		manager.SendMessageUnicast(msg, mac)
		return nil
	}
	return &NoConnectionEstablished{}
}

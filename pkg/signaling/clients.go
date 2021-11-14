package signaling

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	api "github.com/alphahorizon/libentangle/pkg/api/websockets/v1"
	"github.com/google/uuid"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var (
	result []byte
)

type SignalingClient struct {
	onAcceptance   func(conn *websocket.Conn, uuid string) error
	onIntroduction func(conn *websocket.Conn, data []byte, uuid string) error
	onOffer        func(conn *websocket.Conn, data []byte, candidates *chan string, wg *sync.WaitGroup, uuid string) error
	onAnswer       func(data []byte, candidates *chan string, wg *sync.WaitGroup) error
	onCandidate    func(data []byte, candidates *chan string) error
	onResignation  func() error
}

func NewSignalingClient(
	onAcceptance func(conn *websocket.Conn, uuid string) error,
	onIntroduction func(conn *websocket.Conn, data []byte, uuid string) error,
	onOffer func(conn *websocket.Conn, data []byte, candidates *chan string, wg *sync.WaitGroup, uuid string) error,
	onAnswer func(data []byte, candidates *chan string, wg *sync.WaitGroup) error,
	onCandidate func(data []byte, candidates *chan string) error,
	onResignation func() error,
) *SignalingClient {
	return &SignalingClient{
		onAcceptance:   onAcceptance,
		onIntroduction: onIntroduction,
		onOffer:        onOffer,
		onAnswer:       onAnswer,
		onCandidate:    onCandidate,
		onResignation:  onResignation,
	}
}

func (s *SignalingClient) HandleConn(laddrKey string, communityKey string) []byte {
	uuid := uuid.NewString()

	wsAddress := "ws://" + laddrKey
	conn, _, error := websocket.Dial(context.Background(), wsAddress, nil)
	if error != nil {
		log.Fatal(error)
	}
	defer conn.Close(websocket.StatusNormalClosure, "Closing websocket connection nominally")

	var wg sync.WaitGroup
	wg.Add(1)

	candidates := make(chan string)

	go func() {
		if err := wsjson.Write(context.Background(), conn, api.NewApplication(communityKey, uuid)); err != nil {
			log.Fatal(err)
		}

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c

			if err := wsjson.Write(context.Background(), conn, api.NewExited(uuid)); err != nil {
				log.Fatal(err)
			}

			return
		}()

	}()

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
			case api.OpcodeAcceptance:
				s.onAcceptance(conn, uuid)
				break
			case api.OpcodeIntroduction:
				s.onIntroduction(conn, data, uuid)
				break
			case api.OpcodeOffer:
				s.onOffer(conn, data, &candidates, &wg, uuid)
				break
			case api.OpcodeAnswer:
				s.onAnswer(data, &candidates, &wg)
				break
			case api.OpcodeCandidate:
				s.onCandidate(data, &candidates)
				break
			case api.OpcodeResignation:
				s.onResignation()
			}
		}
	}()
	<-exitClient
	if err := wsjson.Write(context.Background(), conn, api.NewExited(uuid)); err != nil {
		log.Fatal(err)
	}
	return result
}

func refString(s string) *string {
	return &s
}

func refUint16(i uint16) *uint16 {
	return &i
}

package signaling

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	api "github.com/alphahorizon/libentangle/pkg/api/websockets/v1"
	"nhooyr.io/websocket"
)

// TODO: modify the signaling protocol
// The signaling protocol is located at /docs/signaling-protocol.txt

type SignalingServer struct {
	onApplication func(application api.Application, conn *websocket.Conn) error
	onReady       func(ready api.Ready, conn *websocket.Conn) error
	onOffer       func(offer api.Offer) error
	onAnswer      func(answer api.Answer) error
	onCandidate   func(candidate api.Candidate) error
	onExited      func(exited api.Exited) error
}

func NewSignalingServer(
	onApplication func(application api.Application, conn *websocket.Conn) error,
	onReady func(ready api.Ready, conn *websocket.Conn) error,
	onOffer func(offer api.Offer) error,
	onAnswer func(answer api.Answer) error,
	onCandidate func(candidate api.Candidate) error,
	onExited func(exited api.Exited) error,
) *SignalingServer {
	return &SignalingServer{
		onApplication: onApplication,
		onReady:       onReady,
		onOffer:       onOffer,
		onAnswer:      onAnswer,
		onCandidate:   onCandidate,
		onExited:      onExited,
	}
}

func (s *SignalingServer) HandleConn(conn websocket.Conn) {

	go func() {
	loop:
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
			case api.OpcodeApplication:
				var application api.Application
				if err := json.Unmarshal(data, &application); err != nil {
					log.Fatal(err)
				}
				s.onApplication(application, &conn)
				break
			case api.OpcodeReady:
				var ready api.Ready
				if err := json.Unmarshal(data, &ready); err != nil {
					log.Fatal(err)
				}
				s.onReady(ready, &conn)
				break
			case api.OpcodeOffer:
				var offer api.Offer
				if err := json.Unmarshal(data, &offer); err != nil {
					log.Fatal(err)
				}
				s.onOffer(offer)
				break
			case api.OpcodeAnswer:
				var answer api.Answer
				if err := json.Unmarshal(data, &answer); err != nil {
					log.Fatal(err)
				}
				s.onAnswer(answer)
				break
			case api.OpcodeCandidate:
				var candidate api.Candidate
				if err := json.Unmarshal(data, &candidate); err != nil {
					log.Fatal(err)
				}
				s.onCandidate(candidate)
				break
			case api.OpcodeExited:
				var exited api.Exited
				if err := json.Unmarshal(data, &exited); err != nil {
					log.Fatal(err)
				}
				s.onExited(exited)
				break loop
			default:
				log.Fatal("Invalid message. Consider using a valid opcode.")
			}
		}
	}()
}

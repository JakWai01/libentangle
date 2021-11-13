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
	"github.com/pion/webrtc/v3"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

var (
	result []byte
)

type SignalingClient struct {
	onAcceptance   func(conn *websocket.Conn, uuid string) error
	onIntroduction func(conn *websocket.Conn, data []byte, peerConnecton *webrtc.PeerConnection) error
	onOffer        func(conn *websocket.Conn, data []byte, peerConnection *webrtc.PeerConnection, candidates *chan string, wg *sync.WaitGroup) error
	onAnswer       func(data []byte, peerConnection *webrtc.PeerConnection, candidates *chan string, wg *sync.WaitGroup) error
	onCandidate    func(data []byte, candidates *chan string) error
	onResignation  func() error
}

func NewSignalingClient(
	onAcceptance func(conn *websocket.Conn, uuid string) error,
	onIntroduction func(conn *websocket.Conn, data []byte, peerConnecton *webrtc.PeerConnection) error,
	onOffer func(conn *websocket.Conn, data []byte, peerConnection *webrtc.PeerConnection, candidates *chan string, wg *sync.WaitGroup) error,
	onAnswer func(data []byte, peerConnection *webrtc.PeerConnection, candidates *chan string, wg *sync.WaitGroup) error,
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
	// The new arguments we pass

	uuid := uuid.NewString()

	// cm := ClientManager{}

	wsAddress := "ws://" + laddrKey
	conn, _, error := websocket.Dial(context.Background(), wsAddress, nil)
	if error != nil {
		log.Fatal(error)
	}
	defer conn.Close(websocket.StatusNormalClosure, "Closing websocket connection nominally")

	// Prepare configuration
	var config = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create RTCPeerConnection
	var peerConnection, err = webrtc.NewPeerConnection(config)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if cErr := peerConnection.Close(); cErr != nil {
			log.Printf("cannot close peerConnection: %v\n", cErr)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	candidates := make(chan string)

	var candidatesMux sync.Mutex

	// Introduce pending candidates. When a remote description is not set yet, the candidates will be cached
	// until a later invocation of the function
	pendingCandidates := make([]*webrtc.ICECandidate, 0)

	// Set the handler for peer connection state
	// This will notify you when the peer has connected/disconnected
	go func() {
		peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
			log.Printf("Peer Connection State has changed: %s\n", s.String())
		})

		// This triggers when WE have a candidate for the other peer, not the other way around
		// This candidate key needs to be send to the other peer
		peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
			fmt.Println("Candidate was generated!")
			if i == nil {
				return
			} else {
				// wg.Wait()
				candidatesMux.Lock()
				defer func() {
					candidatesMux.Unlock()
				}()

				desc := peerConnection.RemoteDescription()

				if desc == nil {
					pendingCandidates = append(pendingCandidates, i)
				} else if err := wsjson.Write(context.Background(), conn, api.NewCandidate(uuid, []byte(i.ToJSON().Candidate))); err != nil {
					log.Fatal(err)
				}
			}

		})

		// Register data channel creation handling
		peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
			// Register channel opening handling
			d.OnOpen(func() {
				if sendErr := d.Send([]byte("Hello World!")); sendErr != nil {
					log.Fatal(sendErr)
				}
			})

		})

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
				fmt.Println(peerConnection.ConnectionState())
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
				// cm.HandleAcceptance(conn, uuid)
				s.onAcceptance(conn, uuid)
				break
			case api.OpcodeIntroduction:
				// cm.HandleIntroduction(conn, data, peerConnection)
				s.onIntroduction(conn, data, peerConnection)
				break
			case api.OpcodeOffer:
				// cm.HandleOffer(conn, data, peerConnection, &candidates, &wg)
				s.onOffer(conn, data, peerConnection, &candidates, &wg)
				break
			case api.OpcodeAnswer:
				// cm.HandleAnswer(data, peerConnection, &candidates, &wg)
				s.onAnswer(data, peerConnection, &candidates, &wg)
				break
			case api.OpcodeCandidate:
				// cm.HandleCandidate(data, &candidates)
				s.onCandidate(data, &candidates)
				break
			case api.OpcodeResignation:
				// cm.HandleResignation()
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

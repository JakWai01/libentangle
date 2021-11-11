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

type SignalingClient struct{}

func NewSignalingClient() *SignalingClient {
	return &SignalingClient{}
}

func (s *SignalingClient) HandleConn(laddrKey string, communityKey string, filename string, file []byte) []byte {
	// The new arguments we pass
	fmt.Println(filename)
	fmt.Println(string(file))

	uuid := uuid.NewString()

	cm := ClientManager{}

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
				cm.HandleAcceptance(conn, uuid)
				break
			case api.OpcodeIntroduction:
				cm.HandleIntroduction(conn, data, peerConnection)
				break
			case api.OpcodeOffer:
				var offer api.Offer
				if err := json.Unmarshal(data, &offer); err != nil {
					log.Fatal(err)
				}

				partnerMac := offer.Mac

				var offer_val webrtc.SessionDescription

				if err := json.Unmarshal([]byte(offer.Payload), &offer_val); err != nil {
					log.Fatal(err)
				}

				if err := peerConnection.SetRemoteDescription(offer_val); err != nil {
					log.Fatal(err)
				}

				go func() {
					for candidate := range candidates {
						if err := peerConnection.AddICECandidate(webrtc.ICECandidateInit{Candidate: candidate, SDPMid: refString("0"), SDPMLineIndex: refUint16(0)}); err != nil {
							log.Fatal(err)
						}
					}
				}()

				answer_val, err := peerConnection.CreateAnswer(nil)
				if err != nil {
					log.Fatal(err)
				}

				err = peerConnection.SetLocalDescription(answer_val)
				if err != nil {
					log.Fatal(err)
				}

				data, err := json.Marshal(answer_val)
				if err != nil {
					log.Fatal(err)
				}

				if err := wsjson.Write(context.Background(), conn, api.NewAnswer(data, partnerMac)); err != nil {
					log.Fatal(err)
				}

				wg.Done()
				break
			case api.OpcodeAnswer:
				var answer api.Answer
				if err := json.Unmarshal(data, &answer); err != nil {
					log.Fatal(err)
				}

				var answer_val webrtc.SessionDescription

				if err := json.Unmarshal([]byte(answer.Payload), &answer_val); err != nil {
					log.Fatal(err)
				}

				if err := peerConnection.SetRemoteDescription(answer_val); err != nil {
					log.Fatal(err)
				}

				go func() {
					for candidate := range candidates {
						if err := peerConnection.AddICECandidate(webrtc.ICECandidateInit{Candidate: candidate, SDPMid: refString("0"), SDPMLineIndex: refUint16(0)}); err != nil {
							log.Fatal(err)
						}
					}
				}()

				wg.Done()
				break
			case api.OpcodeCandidate:
				fmt.Println("received Candidate")
				var candidate api.Candidate
				if err := json.Unmarshal(data, &candidate); err != nil {
					log.Fatal(err)
				}
				fmt.Println(string(candidate.Payload))
				go func() {
					candidates <- string(candidate.Payload)
				}()

				break
			case api.OpcodeResignation:
				exitClient <- struct{}{}
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

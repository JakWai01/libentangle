package signaling

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"

	api "github.com/alphahorizon/libentangle/pkg/api/websockets/v1"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// The signaling protocol is located at /docs/signaling-protocol.txt

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

				if _, ok := s.macs[application.Mac]; ok {
					// Send rejection. That mac is already contained

					// Check if this conn is correct
					if err := wsjson.Write(context.Background(), &conn, api.NewRejection()); err != nil {
						log.Fatal(err)
					}
					break
				}

				s.connections[application.Mac] = conn

				// Check if community exists and if there are less than 2 members inside
				if val, ok := s.communities[application.Community]; ok {
					if len(val) >= 2 {
						// Send rejection. This community is full
						if err := wsjson.Write(context.Background(), &conn, api.NewRejection()); err != nil {
							log.Fatal(err)
						}

						break
					} else {
						// Community exists and has less than 2 members inside
						s.communities[application.Community] = append(s.communities[application.Community], application.Mac)
						s.macs[application.Mac] = false

						if err := wsjson.Write(context.Background(), &conn, api.NewAcceptance()); err != nil {
							log.Fatal(err)
						}

						break
					}
				} else {
					// Community does not exist. Create commuity and insert mac
					s.communities[application.Community] = append(s.communities[application.Community], application.Mac)
					s.macs[application.Mac] = false

					if err := wsjson.Write(context.Background(), &conn, api.NewAcceptance()); err != nil {
						log.Fatal(err)
					}
					break
				}
			case api.OpcodeReady:
				var ready api.Ready
				if err := json.Unmarshal(data, &ready); err != nil {
					log.Fatal(err)
				}

				// If we receive ready, mark the sending person as ready and check if bot hare ready.
				s.macs[ready.Mac] = true

				// Loop through all members of the community and through all elements in it. If the mac isn't member of a community, this will log.Fatal.
				community, err := s.getCommunity(ready.Mac)
				if err != nil {
					log.Fatal(err)
				}

				if len(s.communities[community]) == 2 {
					if s.macs[s.communities[community][0]] == true && s.macs[s.communities[community][1]] == true {
						// Send an introduction to the peer containing the address of the first peer.
						if err := wsjson.Write(context.Background(), &conn, api.NewIntroduction(s.communities[community][0])); err != nil {
							log.Fatal(err)
						}
						break
					}
				}
				break
			case api.OpcodeOffer:
				var offer api.Offer
				if err := json.Unmarshal(data, &offer); err != nil {
					log.Fatal(err)
				}

				// Get the connection of the receiver and send him the payload
				receiver := s.connections[offer.Mac]

				community, err := s.getCommunity(offer.Mac)
				if err != nil {
					log.Fatal(err)
				}

				// We need to assign this
				offer.Mac = s.getSenderMac(offer.Mac, community)

				if err := wsjson.Write(context.Background(), &receiver, offer); err != nil {
					log.Fatal(err)
				}
				break
			case api.OpcodeAnswer:
				var answer api.Answer
				if err := json.Unmarshal(data, &answer); err != nil {
					log.Fatal(err)
				}

				// Get connection of the receiver and send him the payload
				receiver := s.connections[answer.Mac]

				community, err := s.getCommunity(answer.Mac)
				if err != nil {
					log.Fatal(err)
				}

				answer.Mac = s.getSenderMac(answer.Mac, community)

				if err := wsjson.Write(context.Background(), &receiver, answer); err != nil {
					log.Fatal(err)
				}
				break
			case api.OpcodeCandidate:
				var candidate api.Candidate
				if err := json.Unmarshal(data, &candidate); err != nil {
					log.Fatal(err)
				}

				community, err := s.getCommunity(candidate.Mac)
				if err != nil {
					log.Fatal(err)
				}

				candidate.Mac = s.getSenderMac(candidate.Mac, community)

				target := s.connections[candidate.Mac]

				if err := wsjson.Write(context.Background(), &target, candidate); err != nil {
					log.Fatal(err)
				}
				break
			case api.OpcodeExited:
				var exited api.Exited
				if err := json.Unmarshal(data, &exited); err != nil {
					log.Fatal(err)
				}

				var receiver websocket.Conn

				// Get the other peer in the community
				community, err := s.getCommunity(exited.Mac)
				if err != nil {
					log.Fatal(err)
				}

				if len(s.communities[community]) == 2 {
					if exited.Mac == s.communities[community][0] {
						// The second one is receiver
						receiver = s.connections[s.communities[community][1]]
					} else {
						// First one
						receiver = s.connections[s.communities[community][0]]
					}
				} else {
					receiver = s.connections[s.communities[community][0]]
				}

				// Send to the other peer
				if err := wsjson.Write(context.Background(), &receiver, api.NewResignation(exited.Mac)); err != nil {
					log.Fatal(err)
				}

				// Remove this peer from all maps
				delete(s.macs, exited.Mac)
				delete(s.connections, exited.Mac)

				// Remove member from community
				s.communities[community] = deleteElement(s.communities[community], exited.Mac)

				// Remove community only if there is only one member left
				if len(s.communities[community]) == 0 {
					delete(s.communities, community)
				}

				break loop
			default:
				log.Fatal("Invalid message. Consider using a valid opcode.")
			}
		}
	}()
}

func (s *SignalingServer) getCommunity(mac string) (string, error) {
	for key, element := range s.communities {
		for i := 0; i < len(element); i++ {
			if element[i] == mac {
				return key, nil
			}
		}
	}

	return "", errors.New("This mac is not part of any community so far!")
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func deleteElement(s []string, str string) []string {
	var elementIndex int
	for index, element := range s {
		if element == str {
			elementIndex = index
		}
	}
	return append(s[:elementIndex], s[elementIndex+1:]...)
}

func asSha256(o interface{}) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", o)))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (s *SignalingServer) getSenderMac(receiverMac string, community string) string {
	if len(s.communities[community]) == 2 {
		if receiverMac == s.communities[community][1] {
			// The second one is sender
			return s.communities[community][0]
		} else {
			// First one
			return s.communities[community][1]
		}
	} else {
		return s.communities[community][1]
	}
}

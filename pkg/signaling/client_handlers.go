package signaling

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	api "github.com/alphahorizon/libentangle/pkg/api/websockets/v1"
	"github.com/pion/webrtc/v3"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type ClientManager struct{}

func NewClientManager() *ClientManager {
	return &ClientManager{}
}

func (m *ClientManager) HandleAcceptance(conn *websocket.Conn, uuid string) error {
	if err := wsjson.Write(context.Background(), conn, api.NewReady(uuid)); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (m *ClientManager) HandleIntroduction(conn *websocket.Conn, data []byte, peerConnection *webrtc.PeerConnection) error {
	// Create DataChannel
	sendChannel, err := peerConnection.CreateDataChannel("foo", nil)
	if err != nil {
		log.Fatal(err)
	}
	sendChannel.OnClose(func() {
		log.Println("sendChannel has closed")
	})
	sendChannel.OnOpen(func() {
		log.Println("sendChannel has opened")
	})
	sendChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		log.Printf("Message from DataChannel %s payload %s", sendChannel.Label(), string(msg.Data))

		defer sendChannel.Close()

		exitClient <- struct{}{}
	})

	var introduction api.Introduction
	if err := json.Unmarshal(data, &introduction); err != nil {
		log.Fatal(err)
	}

	partnerMac := introduction.Mac

	offer, err := peerConnection.CreateOffer(nil)
	if err != nil {
		log.Fatal(err)
	}

	if err := peerConnection.SetLocalDescription(offer); err != nil {
		log.Fatal(err)
	}

	data, err = json.Marshal(offer)
	if err != nil {
		log.Fatal(err)
	}

	if err := wsjson.Write(context.Background(), conn, api.NewOffer(data, partnerMac)); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (m *ClientManager) HandleOffer(conn *websocket.Conn, data []byte, peerConnection *webrtc.PeerConnection, candidates *chan string, wg *sync.WaitGroup) error {
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
		for candidate := range *candidates {
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

	data, err = json.Marshal(answer_val)
	if err != nil {
		log.Fatal(err)
	}

	if err := wsjson.Write(context.Background(), conn, api.NewAnswer(data, partnerMac)); err != nil {
		log.Fatal(err)
	}

	wg.Done()
	return nil
}

func (m *ClientManager) HandleAnswer(data []byte, peerConnection *webrtc.PeerConnection, candidates *chan string, wg *sync.WaitGroup) error {
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
		for candidate := range *candidates {
			if err := peerConnection.AddICECandidate(webrtc.ICECandidateInit{Candidate: candidate, SDPMid: refString("0"), SDPMLineIndex: refUint16(0)}); err != nil {
				log.Fatal(err)
			}
		}
	}()

	wg.Done()
	return nil
}

func (m *ClientManager) HandleCandidate(data []byte, candidates *chan string) error {
	fmt.Println("received Candidate")
	var candidate api.Candidate
	if err := json.Unmarshal(data, &candidate); err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(candidate.Payload))
	go func() {
		*candidates <- string(candidate.Payload)
	}()
	return nil
}

func (m *ClientManager) HandleResignation() error {
	exitClient <- struct{}{}
	return nil
}

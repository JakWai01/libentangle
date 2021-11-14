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

type ClientManager struct {
	lock sync.Mutex

	peers map[string]*peer
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		peers: map[string]*peer{},
	}
}

type peer struct {
	connection *webrtc.PeerConnection
	channel    *webrtc.DataChannel
	candidates []*webrtc.ICECandidate
}

func (m *ClientManager) HandleAcceptance(conn *websocket.Conn, uuid string) error {
	if err := wsjson.Write(context.Background(), conn, api.NewReady(uuid)); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (m *ClientManager) HandleIntroduction(conn *websocket.Conn, data []byte, uuid string) error {
	var introduction api.Introduction
	if err := json.Unmarshal(data, &introduction); err != nil {
		log.Fatal(err)
	}

	peerConnection, err := m.createPeer(introduction.Mac, conn, uuid)
	if err != nil {
		return err
	}

	if err := m.createDataChannel(introduction.Mac, peerConnection); err != nil {
		return err
	}

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

	if err := wsjson.Write(context.Background(), conn, api.NewOffer(data, uuid, introduction.Mac)); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (m *ClientManager) HandleOffer(conn *websocket.Conn, data []byte, candidates *chan string, wg *sync.WaitGroup, uuid string) error {
	var offer api.Offer
	if err := json.Unmarshal(data, &offer); err != nil {
		log.Fatal(err)
	}

	var offer_val webrtc.SessionDescription

	if err := json.Unmarshal([]byte(offer.Payload), &offer_val); err != nil {
		log.Fatal(err)
	}

	peerConnection, err := m.getPeerConnection(offer.SenderMac)
	if err != nil {
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

	if err := wsjson.Write(context.Background(), conn, api.NewAnswer(data, offer.SenderMac, offer.ReceiverMac)); err != nil {
		log.Fatal(err)
	}

	wg.Done()
	return nil
}

func (m *ClientManager) HandleAnswer(data []byte, candidates *chan string, wg *sync.WaitGroup) error {
	var answer api.Answer
	if err := json.Unmarshal(data, &answer); err != nil {
		log.Fatal(err)
	}

	var answer_val webrtc.SessionDescription

	if err := json.Unmarshal([]byte(answer.Payload), &answer_val); err != nil {
		log.Fatal(err)
	}

	peerConnection, err := m.getPeerConnection(answer.SenderMac)
	if err != nil {
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

func (m *ClientManager) createPeer(mac string, conn *websocket.Conn, uuid string) (*webrtc.PeerConnection, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		log.Printf("Peer Connection State has changed: %s\n", s.String())
	})

	m.peers[mac] = &peer{
		connection: peerConnection,
		candidates: []*webrtc.ICECandidate{},
	}

	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		fmt.Println("Candidate was generated!")
		if i == nil {
			return
		} else {
			if peerConnection.RemoteDescription() == nil {
				m.peers[mac].candidates = append(m.peers[mac].candidates, i)
			} else if err := wsjson.Write(context.Background(), conn, api.NewCandidate([]byte(i.ToJSON().Candidate), uuid, mac)); err != nil {
				log.Fatal(err)
			}
		}
	})

	// Register data channel creation handling
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnOpen(func() {
			if sendErr := d.Send([]byte("Hello World!")); sendErr != nil {
				log.Fatal(sendErr)
			}
		})
	})

	return peerConnection, nil
}

func (m *ClientManager) createDataChannel(mac string, peerConnection *webrtc.PeerConnection) error {
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

	return nil
}

func (m *ClientManager) getPeerConnection(mac string) (*webrtc.PeerConnection, error) {
	return m.peers[mac].connection, nil
}

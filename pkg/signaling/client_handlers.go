package signaling

import (
	"context"
	"encoding/json"
	"log"

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

package signaling

import (
	"context"
	"log"
	"sync"

	api "github.com/alphahorizonio/libentangle/pkg/api/websockets/v1"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type ServerManager struct {
	lock sync.Mutex

	communities map[string][]string
	macs        map[string]websocket.Conn
}

func NewCommunitiesManager() *ServerManager {
	return &ServerManager{
		communities: map[string][]string{},
		macs:        map[string]websocket.Conn{},
	}
}

func (m *ServerManager) HandleApplication(application api.Application, conn *websocket.Conn) error {
	if _, ok := m.macs[application.Mac]; ok {
		// Send rejection. That mac is already contained

		if err := wsjson.Write(context.Background(), conn, api.NewRejection()); err != nil {
			log.Fatal(err)
		}
		return nil
	}

	m.macs[application.Mac] = *conn

	// Check if community exists
	if _, ok := m.communities[application.Community]; ok {
		m.communities[application.Community] = append(m.communities[application.Community], application.Mac)

		if err := wsjson.Write(context.Background(), conn, api.NewAcceptance()); err != nil {
			log.Fatal(err)
		}

		return nil
	} else {
		// Community does not exist. Create commuity and insert mac
		m.communities[application.Community] = append(m.communities[application.Community], application.Mac)

		if err := wsjson.Write(context.Background(), conn, api.NewAcceptance()); err != nil {
			log.Fatal(err)
		}
		return nil
	}
}

func (m *ServerManager) HandleReady(ready api.Ready, conn *websocket.Conn) error {
	community, err := m.getCommunity(ready.Mac)
	if err != nil {
		log.Fatal(err)
	}

	// Broadcast the introduction to all connections, excluding our own
	for _, mac := range m.communities[community] {
		if mac != ready.Mac {
			receiver := m.macs[mac]

			// ensure that ready.Mac == m.communities[community][0]
			if err := wsjson.Write(context.Background(), &receiver, api.NewIntroduction(ready.Mac)); err != nil {
				log.Fatal(err)
			}
		} else {
			continue
		}
	}
	return nil
}

func (m *ServerManager) HandleOffer(offer api.Offer) error {
	receiver := m.macs[offer.ReceiverMac]

	if err := wsjson.Write(context.Background(), &receiver, offer); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (m *ServerManager) HandleAnswer(answer api.Answer) error {
	receiver := m.macs[answer.ReceiverMac]

	if err := wsjson.Write(context.Background(), &receiver, answer); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (m *ServerManager) HandleCandidate(candidate api.Candidate) error {
	receiver := m.macs[candidate.ReceiverMac]

	if err := wsjson.Write(context.Background(), &receiver, candidate); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (m *ServerManager) HandleExited(exited api.Exited) error {
	community, err := m.getCommunity(exited.Mac)
	if err != nil {
		log.Fatal(err)
	}

	for _, mac := range m.communities[community] {
		if mac != exited.Mac {
			receiver := m.macs[mac]

			if err := wsjson.Write(context.Background(), &receiver, api.NewResignation(exited.Mac)); err != nil {
				log.Fatal(err)
			}
		} else {
			continue
		}
	}

	// Remove this peer from all maps
	delete(m.macs, exited.Mac)
	delete(m.macs, exited.Mac)

	// Remove member from community
	m.communities[community] = deleteElement(m.communities[community], exited.Mac)

	// Remove community only if there is only one member left
	if len(m.communities[community]) == 0 {
		delete(m.communities, community)
	}

	return nil
}

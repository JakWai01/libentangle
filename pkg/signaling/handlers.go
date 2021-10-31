package signaling

import (
	"context"
	"log"
	"sync"

	api "github.com/alphahorizon/libentangle/pkg/api/websockets/v1"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

type CommunitiesManager struct {
	lock sync.Mutex

	communities map[string][]string
	macs        map[string]bool
	connections map[string]websocket.Conn
}

func NewCommunitiesManager() *CommunitiesManager {
	return &CommunitiesManager{
		communities: map[string][]string{},
		macs:        map[string]bool{},
		connections: map[string]websocket.Conn{},
	}
}

func (m *CommunitiesManager) HandleApplication(application api.Application, conn *websocket.Conn) error {
	if _, ok := m.macs[application.Mac]; ok {
		// Send rejection. That mac is already contained

		// Check if this conn is correct
		if err := wsjson.Write(context.Background(), conn, api.NewRejection()); err != nil {
			log.Fatal(err)
		}
		return nil
	}

	m.connections[application.Mac] = *conn

	// Check if community exists and if there are less than 2 members inside
	if val, ok := m.communities[application.Community]; ok {
		if len(val) >= 2 {
			// Send rejection. This community is full
			if err := wsjson.Write(context.Background(), conn, api.NewRejection()); err != nil {
				log.Fatal(err)
			}

			return nil
		} else {
			// Community exists and has less than 2 members inside
			m.communities[application.Community] = append(m.communities[application.Community], application.Mac)
			m.macs[application.Mac] = false

			if err := wsjson.Write(context.Background(), conn, api.NewAcceptance()); err != nil {
				log.Fatal(err)
			}

			return nil
		}
	} else {
		// Community does not exist. Create commuity and insert mac
		m.communities[application.Community] = append(m.communities[application.Community], application.Mac)
		m.macs[application.Mac] = false

		if err := wsjson.Write(context.Background(), conn, api.NewAcceptance()); err != nil {
			log.Fatal(err)
		}
		return nil
	}
}

func (m *CommunitiesManager) HandleReady(ready api.Ready, conn *websocket.Conn) error {
	// If we receive ready, mark the sending person as ready and check if bot hare ready.
	m.macs[ready.Mac] = true

	// Loop through all members of the community and through all elements in it. If the mac isn't member of a community, this will log.Fatal.
	community, err := m.getCommunity(ready.Mac)
	if err != nil {
		log.Fatal(err)
	}

	if len(m.communities[community]) == 2 {
		if m.macs[m.communities[community][0]] == true && m.macs[m.communities[community][1]] == true {
			// Send an introduction to the peer containing the address of the first peer.
			if err := wsjson.Write(context.Background(), conn, api.NewIntroduction(m.communities[community][0])); err != nil {
				log.Fatal(err)
			}
			return nil
		}
	}
	return nil
}

func (m *CommunitiesManager) HandleOffer(offer api.Offer) error {
	// Get the connection of the receiver and send him the payload
	receiver := m.connections[offer.Mac]

	community, err := m.getCommunity(offer.Mac)
	if err != nil {
		log.Fatal(err)
	}

	// We need to assign this
	offer.Mac = m.getSenderMac(offer.Mac, community)

	if err := wsjson.Write(context.Background(), &receiver, offer); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (m *CommunitiesManager) HandleAnswer(answer api.Answer) error {
	// Get connection of the receiver and send him the payload
	receiver := m.connections[answer.Mac]

	community, err := m.getCommunity(answer.Mac)
	if err != nil {
		log.Fatal(err)
	}

	answer.Mac = m.getSenderMac(answer.Mac, community)

	if err := wsjson.Write(context.Background(), &receiver, answer); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (m *CommunitiesManager) HandleCandidate(candidate api.Candidate) error {
	community, err := m.getCommunity(candidate.Mac)
	if err != nil {
		log.Fatal(err)
	}

	candidate.Mac = m.getSenderMac(candidate.Mac, community)

	target := m.connections[candidate.Mac]

	if err := wsjson.Write(context.Background(), &target, candidate); err != nil {
		log.Fatal(err)
	}
	return nil
}

func (m *CommunitiesManager) HandleExited(exited api.Exited) error {
	var receiver websocket.Conn

	// Get the other peer in the community
	community, err := m.getCommunity(exited.Mac)
	if err != nil {
		log.Fatal(err)
	}

	if len(m.communities[community]) == 2 {
		if exited.Mac == m.communities[community][0] {
			// The second one is receiver
			receiver = m.connections[m.communities[community][1]]
		} else {
			// First one
			receiver = m.connections[m.communities[community][0]]
		}
	} else {
		receiver = m.connections[m.communities[community][0]]
	}

	// Send to the other peer
	if err := wsjson.Write(context.Background(), &receiver, api.NewResignation(exited.Mac)); err != nil {
		log.Fatal(err)
	}

	// Remove this peer from all maps
	delete(m.macs, exited.Mac)
	delete(m.connections, exited.Mac)

	// Remove member from community
	m.communities[community] = deleteElement(m.communities[community], exited.Mac)

	// Remove community only if there is only one member left
	if len(m.communities[community]) == 0 {
		delete(m.communities, community)
	}

	return nil
}

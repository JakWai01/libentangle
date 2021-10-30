package api

type Application struct {
	Message
	Community string `json:"community"`
	Mac       string `json:"mac"`
}

type Acceptance struct {
	Message
}

type Rejection struct {
	Message
}

type Ready struct {
	Message
	Mac string `json:"mac"`
}

type Introduction struct {
	Message
	Mac string `json:"mac"`
}

type Offer struct {
	Message
	Payload []byte `json:"payload"`
	Mac     string `json:"mac"`
}

type Answer struct {
	Message
	Payload []byte `json:"payload"`
	Mac     string `json:"mac"`
}

type Candidate struct {
	Message
	Payload []byte `json:"payload"`
	Mac     string `json:"mac"`
}

type Exited struct {
	Message
	Mac string `json:"mac"`
}

type Resignation struct {
	Message
	Mac string `json:"mac"`
}

func NewApplication(community string, mac string) *Application {
	return &Application{Message: Message{OpcodeApplication}, Community: community, Mac: mac}
}

func NewAcceptance() *Acceptance {
	return &Acceptance{Message: Message{OpcodeAcceptance}}
}

func NewRejection() *Rejection {
	return &Rejection{Message: Message{OpcodeRejection}}
}

func NewReady(mac string) *Ready {
	return &Ready{Message: Message{OpcodeReady}, Mac: mac}
}

func NewIntroduction(mac string) *Introduction {
	return &Introduction{Message: Message{OpcodeIntroduction}, Mac: mac}
}

func NewOffer(payload []byte, mac string) *Offer {
	return &Offer{Message: Message{OpcodeOffer}, Payload: payload, Mac: mac}
}

func NewAnswer(payload []byte, mac string) *Answer {
	return &Answer{Message: Message{OpcodeAnswer}, Payload: payload, Mac: mac}
}

func NewCandidate(mac string, payload []byte) *Candidate {
	return &Candidate{Message: Message{OpcodeCandidate}, Payload: payload, Mac: mac}
}

func NewExited(mac string) *Exited {
	return &Exited{Message: Message{OpcodeExited}, Mac: mac}
}

func NewResignation(mac string) *Resignation {
	return &Resignation{Message: Message{OpcodeResignation}, Mac: mac}
}

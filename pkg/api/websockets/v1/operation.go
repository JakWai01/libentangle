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
	Payload     []byte `json:"payload"`
	SenderMac   string `json:"sender"`
	ReceiverMac string `json:"receiver"`
}

type Answer struct {
	Message
	Payload     []byte `json:"payload"`
	SenderMac   string `json:"sender"`
	ReceiverMac string `json:"receiver"`
}

type Candidate struct {
	Message
	Payload     []byte `json:"payload"`
	SenderMac   string `json:"sender"`
	ReceiverMac string `json:"receiver"`
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

func NewOffer(payload []byte, sender string, receiver string) *Offer {
	return &Offer{Message: Message{OpcodeOffer}, Payload: payload, SenderMac: sender, ReceiverMac: receiver}
}

func NewAnswer(payload []byte, sender string, receiver string) *Answer {
	return &Answer{Message: Message{OpcodeAnswer}, Payload: payload, SenderMac: sender, ReceiverMac: receiver}
}

func NewCandidate(payload []byte, sender string, receiver string) *Candidate {
	return &Candidate{Message: Message{OpcodeCandidate}, Payload: payload, SenderMac: sender, ReceiverMac: receiver}
}

func NewExited(mac string) *Exited {
	return &Exited{Message: Message{OpcodeExited}, Mac: mac}
}

func NewResignation(mac string) *Resignation {
	return &Resignation{Message: Message{OpcodeResignation}, Mac: mac}
}

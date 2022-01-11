package api

type Message struct {
	Opcode string `json:"opcode"`
}

type WrappedMessage struct {
	Mac     string `json:"mac"`
	Payload []byte `json:"payload"`
}

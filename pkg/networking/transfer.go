package networking

type Message struct {
	Name    string `json:"name"`
	Content []byte `json:"content"`
}

func Write(p []byte) error {
	if manager != nil {
		manager.SendMessage(p)
		return nil
	}
	return &NoConnectionEstablished{}
}

func WriteUnicast(p []byte, mac string) error {
	if manager != nil {
		manager.SendMessageUnicast(p, mac)
		return nil
	}
	return &NoConnectionEstablished{}
}

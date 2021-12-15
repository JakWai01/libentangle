package networking

type Message struct {
	Name    string `json:"name"`
	Content []byte `json:"content"`
}

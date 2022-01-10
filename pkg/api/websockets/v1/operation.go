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

type ReadOp struct {
	Message
	Length int `json:"length"`
}

type WriteOp struct {
	Message
	Payload []byte `json:"payload"`
}

type SeekOp struct {
	Message
	Offset int64 `json:"offset"`
	Whence int   `json:"whence"`
}

type OpenOp struct {
	Message
	ReadOnly bool
}

type CloseOp struct {
	Message
}

type ReadOpResponse struct {
	Message
	Bytes     []byte `json:"bytes"`
	BytesRead int64  `json:"bytesread"`
	Error     string `json:"error"`
}

type WriteOpResponse struct {
	Message
	BytesRead int64  `json:"bytesread"`
	Error     string `json:"error"`
}

type SeekOpResponse struct {
	Message
	Offset int64  `json:"offset"`
	Error  string `json:"error"`
}

// We are missing a field here
type OpenOpResponse struct {
	Message
	Error string `json:"error"`
}

type CloseOpResponse struct {
	Message
	Error string `json:"error"`
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

func NewReadOp(length int) *ReadOp {
	return &ReadOp{Message: Message{OpcodeRead}, Length: length}
}

func NewWriteOp(payload []byte) *WriteOp {
	return &WriteOp{Message: Message{OpcodeWrite}, Payload: payload}
}

func NewSeekOp(offset int64, whence int) *SeekOp {
	return &SeekOp{Message: Message{OpcodeSeek}, Offset: offset, Whence: whence}
}

func NewOpenOp(readOnly bool) *OpenOp {
	return &OpenOp{Message: Message{OpcodeOpen}, ReadOnly: readOnly}
}

func NewCloseOp() *CloseOp {
	return &CloseOp{Message: Message{OpcodeClose}}
}

func NewReadOpResponse(bytes []byte, bytesread int64, err string) *ReadOpResponse {
	return &ReadOpResponse{Message: Message{OpcodeReadResponse}, Bytes: bytes, BytesRead: bytesread, Error: err}
}

func NewWriteOpResponse(bytesread int64, err string) *WriteOpResponse {
	return &WriteOpResponse{Message: Message{OpcodeWriteResponse}, BytesRead: bytesread, Error: err}
}

func NewSeekOpResponse(offset int64, err string) *SeekOpResponse {
	return &SeekOpResponse{Message: Message{OpcodeSeekResponse}, Offset: offset, Error: err}
}

func NewOpenOpResponse(err string) *OpenOpResponse {
	return &OpenOpResponse{Message: Message{OpcodeOpenResponse}, Error: err}
}

func NewCloseOpResponse(err string) *CloseOpResponse {
	return &CloseOpResponse{Message: Message{OpcodeCloseResponse}, Error: err}
}

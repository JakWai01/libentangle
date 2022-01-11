package api

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

type OpenOpResponse struct {
	Message
	Error string `json:"error"`
}

type CloseOpResponse struct {
	Message
	Error string `json:"error"`
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

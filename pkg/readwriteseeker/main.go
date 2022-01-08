package readwriteseeker

import (
	"encoding/json"

	api "github.com/alphahorizonio/libentangle/pkg/api/websockets/v1"
	"github.com/alphahorizonio/libentangle/pkg/networking"
)

var ()

type FileSystemError struct {
	err string
}

func (e *FileSystemError) Error() string {
	return e.err
}

type RemoteFile struct {
	ReadCh  chan api.ReadOpResponse
	WriteCh chan api.WriteOpResponse
	SeekCh  chan api.SeekOpResponse
	CloseCh chan api.CloseOpResponse
}

var OpenCh chan api.OpenOpResponse

func Open(name string) (*RemoteFile, error) {

	msg, err := json.Marshal(api.NewOpenOp(name))
	if err != nil {
		panic(err)
	}

	networking.WriteToDataChannel(msg)

	response := <-OpenCh

	return &RemoteFile{
		ReadCh:  make(chan api.ReadOpResponse),
		WriteCh: make(chan api.WriteOpResponse),
		SeekCh:  make(chan api.SeekOpResponse),
		CloseCh: make(chan api.CloseOpResponse),
	}, checkError(response.Error)
}

func (f *RemoteFile) Close() error {

	msg, err := json.Marshal(api.NewCloseOp())
	if err != nil {
		panic(err)
	}

	networking.WriteToDataChannel(msg)

	response := <-f.CloseCh

	return checkError(response.Error)
}

// .Connect() responding to a channel onOpen

func (f *RemoteFile) Read(n []byte) (int, error) {

	msg, err := json.Marshal(api.NewReadOp(len(n)))
	if err != nil {
		panic(err)
	}

	networking.WriteToDataChannel(msg)

	response := <-f.ReadCh

	copy(n, response.Bytes)

	return int(response.BytesRead), checkError(response.Error)
}

func (f *RemoteFile) Write(n []byte) (int, error) {

	msg, err := json.Marshal(api.NewWriteOp(n))
	if err != nil {
		panic(err)
	}

	networking.WriteToDataChannel(msg)

	response := <-f.WriteCh

	return int(response.BytesRead), checkError(response.Error)
}

func (f *RemoteFile) Seek(offset int64, whence int) (int64, error) {

	msg, err := json.Marshal(api.NewSeekOp(offset, whence))
	if err != nil {
		panic(err)
	}

	networking.WriteToDataChannel(msg)

	response := <-f.SeekCh

	return response.Offset, checkError(response.Error)
}

func checkError(err string) error {
	switch err {
	case NoneKey:
		return nil
	default:
		return &FileSystemError{err: err}
	}
}

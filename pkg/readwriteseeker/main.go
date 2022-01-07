package readwriteseeker

import (
	"encoding/json"

	api "github.com/alphahorizonio/libentangle/pkg/api/websockets/v1"
	"github.com/alphahorizonio/libentangle/pkg/networking"
)

var (
	readCh  chan api.ReadOpResponse
	writeCh chan api.WriteOpResponse
	seekCh  chan api.SeekOpResponse
)

type Writer interface {
	Write([]byte) (int, error)
}

type Reader interface {
	Read([]byte) (int, error)
}

type Seeker interface {
	Seek(int64, int) (int64, error)
}

type ReadWriteSeeker interface {
	Reader
	Writer
	Seeker
}

type FileSystemError struct {
	err string
}

func (e *FileSystemError) Error() string {
	return e.err
}

// The connection could not be established yet when the first functioncalls happen

// These are the client functions
// The client dataChannel receives the appropriate response and writes them to a channel where the appropriate method below reads from it
// Have a channel using ReadOpResponse, WriteOpResponse and SeekOpResponse
func Read(n []byte) (int, error) {

	msg, err := json.Marshal(api.NewReadOp(n))
	if err != nil {
		panic(err)
	}

	networking.WriteToDataChannel(msg)

	response := <-readCh

	return response.BytesRead, checkError(response.Error)
}

func Write(n []byte) (int, error) {

	msg, err := json.Marshal(api.NewWriteOp(n))
	if err != nil {
		panic(err)
	}

	networking.WriteToDataChannel(msg)

	response := <-writeCh

	return int(response.BytesRead), checkError(response.Error)
}

func Seek(offset int64, whence int) (int64, error) {

	msg, err := json.Marshal(api.NewSeekOp(offset, whence))
	if err != nil {
		panic(err)
	}

	networking.WriteToDataChannel(msg)

	response := <-seekCh

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

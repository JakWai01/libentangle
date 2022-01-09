package readwriteseeker

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"runtime/debug"
	"sync"

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
	lock    sync.Mutex
	opened  bool
	oplock  sync.Mutex
	ReadCh  chan api.ReadOpResponse
	WriteCh chan api.WriteOpResponse
	SeekCh  chan api.SeekOpResponse
	CloseCh chan api.CloseOpResponse
	OpenCh  chan api.OpenOpResponse
}

func NewRemoteFile() *RemoteFile {
	return &RemoteFile{
		ReadCh:  make(chan api.ReadOpResponse),
		WriteCh: make(chan api.WriteOpResponse),
		SeekCh:  make(chan api.SeekOpResponse),
		CloseCh: make(chan api.CloseOpResponse),
		OpenCh:  make(chan api.OpenOpResponse),
	}
}

func (f *RemoteFile) Fd() uintptr {
	return 1
}

// Is this doing the right thing
func (f *RemoteFile) Open(create bool) error {
	log.Println("REMOTEFILE.Open", create, string(debug.Stack()))
	f.oplock.Lock()
	defer f.oplock.Unlock()

	if f.opened == true {
		return nil
	}

	log.Println("LOCKING")
	f.lock.Lock()
	log.Println("AFTER LOCKING")

	msg, err := json.Marshal(api.NewOpenOp(create))
	if err != nil {
		panic(err)
	}

	networking.WriteToDataChannel(msg)

	log.Println("ASLDKJASLKDJ")
	errorChan := make(chan string)
	go func() {
		err := <-f.OpenCh
		log.Println("USA FUCKS", err)
		errorChan <- err.Error
	}()

	error2 := <-errorChan

	f.opened = true

	if error2 == "" {
		return nil
	}

	return errors.New(error2)
}

func (f *RemoteFile) Close() error {
	log.Println("REMOTEFILE.CLose", string(debug.Stack()))

	f.oplock.Lock()
	defer f.oplock.Unlock()

	log.Println("CLOSING FILE")
	msg, err := json.Marshal(api.NewCloseOp())
	if err != nil {
		panic(err)
	}

	networking.WriteToDataChannel(msg)

	response := <-f.CloseCh

	log.Println("UNLOCKING")
	if f.opened != false {
		f.lock.Unlock()
	}
	log.Println("AFTER UNLOCKING")
	f.opened = false
	return getError(response.Error)
}

// .Connect() responding to a channel onOpen

func (f *RemoteFile) Read(n []byte) (int, error) {
	f.oplock.Lock()
	defer f.oplock.Unlock()
	msg, err := json.Marshal(api.NewReadOp(len(n)))
	if err != nil {
		panic(err)
	}

	networking.WriteToDataChannel(msg)

	response := <-f.ReadCh

	copy(n, response.Bytes)

	return int(response.BytesRead), getError(response.Error)
}

func (f *RemoteFile) Write(n []byte) (int, error) {
	f.oplock.Lock()
	defer f.oplock.Unlock()
	msg, err := json.Marshal(api.NewWriteOp(n))
	if err != nil {
		panic(err)
	}

	networking.WriteToDataChannel(msg)

	response := <-f.WriteCh

	return int(response.BytesRead), getError(response.Error)
}

func (f *RemoteFile) Seek(offset int64, whence int) (int64, error) {
	f.oplock.Lock()
	defer f.oplock.Unlock()
	msg, err := json.Marshal(api.NewSeekOp(offset, whence))
	if err != nil {
		panic(err)
	}

	networking.WriteToDataChannel(msg)

	response := <-f.SeekCh

	return response.Offset, getError(response.Error)
}

func getError(err string) error {
	switch err {
	case NoneKey:
		return nil
	case "EOF":
		return io.EOF
	default:
		return errors.New(err)
	}
}

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"syscall"

	"github.com/alphahorizon/libentangle/pkg/networking"
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"github.com/pion/webrtc/v3"
)

// either write all files and folder to arrays and add them then or add them one by one
type Message struct {
	Opcode string `json:"opcode"`
}
type File struct {
	Message
	Name    string `json:name`
	Content []byte `json:content`
}

type Folder struct {
	Message
	Name string `json:name`
}

type HelloRoot struct {
	fs.Inode
}

// Handle messages from the datachannel here. Receive some opcode when we are done adding the files.
func (r *HelloRoot) OnAdd(ctx context.Context) {
	ch := r.NewPersistentInode(
		ctx, &fs.MemRegularFile{
			// File content
			Data: []byte("Hello World"),
			Attr: fuse.Attr{
				Mode: 0644,
			},
		}, fs.StableAttr{Ino: 2})
	// File name
	r.AddChild("file.txt", ch, false)
}

func (r *HelloRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0755
	return 0
}

func main() {
	flag.Parse()
	opts := &fs.Options{}

	server, err := fs.Mount(flag.Arg(0), &HelloRoot{}, opts)
	if err != nil {
		log.Fatalf("Mount fail: %v\n", err)
	}
	defer server.Unmount()

	networking.Connect("test", func(msg webrtc.DataChannelMessage) {
		log.Printf("Message: %s", msg.Data)

		var v Message
		if err := json.Unmarshal(msg.Data, &v); err != nil {
			log.Fatal(err)
		}

		switch v.Opcode {
		case "folder":
			fmt.Println("folder")
			var folder Folder
			if err := json.Unmarshal(msg.Data, &folder); err != nil {
				log.Fatal(err)
			}
			fmt.Println(folder.Name)
		case "file":
			fmt.Println("file")
			var file File
			if err := json.Unmarshal(msg.Data, &file); err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(file.Content))
		default:
			log.Fatal("Invalid opcode!")
		}
	})

	server.Wait()
}

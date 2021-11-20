package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
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

// At the moment nothing happens when the file system is initialized
func (r *HelloRoot) OnAdd(ctx context.Context) {
	// ch := r.NewPersistentInode(
	// 	ctx, &fs.MemRegularFile{
	// 		// File content
	// 		Data: []byte("Hello World"),
	// 		Attr: fuse.Attr{
	// 			Mode: 0644,
	// 		},
	// 	}, fs.StableAttr{Ino: 2})
	// // File name
	// r.AddChild("file.txt", ch, false)
}

func (r *HelloRoot) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = 0755
	return 0
}

var _ = (fs.NodeGetattrer)((*HelloRoot)(nil))
var _ = (fs.NodeOnAdder)((*HelloRoot)(nil))

func main() {
	flag.Parse()
	opts := &fs.Options{}

	root := &HelloRoot{}

	server, err := fs.Mount(flag.Arg(0), root, opts)
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

			os.Mkdir(folder.Name, 0777)

		case "file":
			fmt.Println("file")
			var file File
			if err := json.Unmarshal(msg.Data, &file); err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(file.Content))

			ch := root.NewPersistentInode(
				context.Background(), &fs.MemRegularFile{
					// File content
					Data: file.Content,
					Attr: fuse.Attr{
						Mode: 0644,
					},
				}, fs.StableAttr{Ino: 2})
			// File name
			root.AddChild(file.Name, ch, false)
		default:
			log.Fatal("Invalid opcode!")
		}
	})

	server.Wait()
}

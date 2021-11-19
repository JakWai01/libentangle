package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/alphahorizon/libentangle/pkg/networking"
	"github.com/pion/webrtc/v3"
)

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

func main() {
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

	select {}
}

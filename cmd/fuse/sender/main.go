package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/alphahorizon/libentangle/pkg/signaling"
	"nhooyr.io/websocket"
)

type File struct {
	Name    string `json:name`
	Content []byte `json:content`
}

type Folder struct {
	Name string `json:name`
}

func main() {

	manager := signaling.NewClientManager()

	client := signaling.NewSignalingClient(
		func(conn *websocket.Conn, uuid string) error {
			return manager.HandleAcceptance(conn, uuid)
		},
		func(conn *websocket.Conn, data []byte, uuid string, wg *sync.WaitGroup) error {
			return manager.HandleIntroduction(conn, data, uuid, wg)
		},
		func(conn *websocket.Conn, data []byte, candidates *chan string, wg *sync.WaitGroup, uuid string) error {
			return manager.HandleOffer(conn, data, candidates, wg, uuid)
		},
		func(data []byte, candidates *chan string, wg *sync.WaitGroup) error {
			return manager.HandleAnswer(data, candidates, wg)
		},
		func(data []byte, candidates *chan string) error {
			return manager.HandleCandidate(data, candidates)
		},
		func() error {
			return manager.HandleResignation()
		},
	)

	go func() {
		go client.HandleConn("localhost:9090", "test")
	}()

	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')

	files, err := ioutil.ReadDir("../example")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		fmt.Println(f.Name())

		// If File, open it, create File, send it
		// If Folder, create Folder, send it

		if f.IsDir() {
			folder := Folder{Name: f.Name()}
			b, err := json.Marshal(folder)
			if err != nil {
				log.Println(err)
			}
			fmt.Println(string(b))

			// send json to receiver
			manager.SendMessage(string(b))
		} else {
			// Get this path from somewhere else
			if f.Name() != "picture.png" {
				dat, err := os.ReadFile(fmt.Sprintf("../example/%s", f.Name()))
				if err != nil {
					log.Fatal(err)
				}

				file := File{Name: f.Name(), Content: dat}

				b, err := json.Marshal(file)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(string(b))
				manager.SendMessage(string(b))
			}

		}

	}

	for {
	}
}

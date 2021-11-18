package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/alphahorizon/libentangle/pkg/networking"
)

type File struct {
	Name    string `json:name`
	Content []byte `json:content`
}

type Folder struct {
	Name string `json:name`
}

func main() {

	networking.Connect("test")

	// We need some kind of way to wait until we are connected, can we return the Connect function onOpen and actually do the stuff in the background?
	// We can return a datachannel and overwrite the OnMessage function to what we want to do.
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
			networking.Write(string(b))
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
				networking.Write(string(b))
			}

		}

	}

	for {
	}
}

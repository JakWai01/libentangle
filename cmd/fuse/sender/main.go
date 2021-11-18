package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

type File struct {
	Name    string `json:name`
	Content []byte `json:content`
}

type Folder struct {
	Name string `json:name`
}

func main() {
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
		} else {
			// Get this path from somewhere else
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
		}

	}
}

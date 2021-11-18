package main

import (
	"fmt"
	"io/ioutil"
	"log"
)

type File struct {
	name    string
	content []byte
}

type Folder struct {
	name string
}

func main() {
	files, err := ioutil.ReadDir("../example")
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		fmt.Println(f.Name())
	}
}

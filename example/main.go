package main

import (
	"bufio"
	"os"

	"github.com/alphahorizon/libentangle/pkg/networking"
)

func main() {
	networking.Connect("test")

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		networking.Write(text)
	}
}

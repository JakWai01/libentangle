package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/alphahorizon/libentangle/pkg/networking"
	"github.com/alphahorizon/libentangle/pkg/signaling"
)

func main() {
	networking.Connect("test")

	<-signaling.DC
	// <-signaling.DC2
	// dc.OnMessage(func(msg webrtc.DataChannelMessage) {
	// 	log.Printf("Message from Jakobs DataChannel %s payload %s", dc.Label(), string(msg.Data))
	// })
	// dc2.OnMessage(func(msg webrtc.DataChannelMessage) {
	// 	log.Printf("Message from Jakobs DataChannel %s payload %s", dc.Label(), string(msg.Data))
	// })

	fmt.Println("DAS:LDKSA:LKD:SAKD:LA")

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		networking.Write([]byte(text))

	}
}

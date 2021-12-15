package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os"

	"github.com/alphahorizonio/libentangle/pkg/networking"
	"github.com/pion/webrtc/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	communityKey = "community"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start a signaling client.",
	RunE: func(cmd *cobra.Command, args []string) error {
		networking.Connect("test", func(msg webrtc.DataChannelMessage) {

			// Call the Reader function which gets the msg.Data each time
			log.Printf("Message: %s", msg.Data)

			var f *os.File
			var err error

			var file networking.Message

			if err = json.Unmarshal(msg.Data, &file); err != nil {
				panic(err)
			}

			f, err = os.OpenFile("output.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}
			defer f.Close()

			if bytes.Equal(file.Content, []byte("BOF")) {
				f.Truncate(0)
				f.Seek(0, 0)
			} else if bytes.Equal(file.Content, []byte("EOF")) {
				return
			} else {
				_, err := f.Write(file.Content)
				if err != nil {
					panic(err)
				}
			}

		})

		for {
			reader := bufio.NewReader(os.Stdin)
			reader.ReadString('\n')

			networking.EntangledWriter("test.txt")

		}
	},
}

func init() {
	clientCmd.PersistentFlags().String(communityKey, "testCommunityName", "Community to join")

	// Bind env variables
	if err := viper.BindPFlags(clientCmd.PersistentFlags()); err != nil {
		log.Fatal("could not bind flags:", err)
	}
	viper.SetEnvPrefix("airdrip")
	viper.AutomaticEnv()
}

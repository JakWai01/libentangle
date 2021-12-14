package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os"

	"github.com/alphahorizon/libentangle/pkg/networking"
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
			log.Printf("Message: %s", msg.Data)

			var f *os.File
			var err error

			type message struct {
				Name    string `json:"name"`
				Content []byte `json:"content"`
			}

			var file message

			if err = json.Unmarshal(msg.Data, &file); err != nil {
				panic(err)
			}

			f, err = os.OpenFile(file.Name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

			// Call own function here
			networking.ReadWriter()

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

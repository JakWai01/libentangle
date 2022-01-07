package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	api "github.com/alphahorizonio/libentangle/pkg/api/websockets/v1"
	"github.com/alphahorizonio/libentangle/pkg/networking"
	"github.com/pion/webrtc/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start a signaling client.",
	RunE: func(cmd *cobra.Command, args []string) error {
		networking.Connect("test", func(msg webrtc.DataChannelMessage) {

			// Handle client messages
			fmt.Println(string(msg.Data))

			var v api.Message

			if err := json.Unmarshal(msg.Data, &v); err != nil {
				panic(err)
			}

			switch v.Opcode {
			case api.OpcodeReadResponse:
				var readOpResponse api.ReadOp
				if err := json.Unmarshal(msg.Data, &readOpResponse); err != nil {
					panic(err)
				}

				break
			case api.OpcodeWriteResponse:
				var writeOpResponse api.WriteOp
				if err := json.Unmarshal(msg.Data, &writeOpResponse); err != nil {
					panic(err)
				}

				break
			case api.OpcodeSeekResponse:
				var seekOpResponse api.SeekOp
				if err := json.Unmarshal(msg.Data, &seekOpResponse); err != nil {
					panic(err)
				}

				break
			}
		})

		// This can be replaced by a select{}, since the messages are sent by the ReadWriteSeeker
		for {
			reader := bufio.NewReader(os.Stdin)
			msg, err := reader.ReadString('\n')
			if err != nil {
				panic(err)
			}

			networking.Write([]byte(msg))
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

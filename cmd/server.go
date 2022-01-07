package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/alphahorizonio/libentangle/pkg/networking"
	"github.com/pion/webrtc/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	api "github.com/alphahorizonio/libentangle/pkg/api/websockets/v1"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start a signaling client acting as a server after connecting",
	RunE: func(cmd *cobra.Command, args []string) error {
		networking.Connect("test", func(msg webrtc.DataChannelMessage) {

			// Handle server messages
			fmt.Println(string(msg.Data))

			var v api.Message

			if err := json.Unmarshal(msg.Data, &v); err != nil {
				panic(err)
			}

			switch v.Opcode {
			case api.OpcodeRead:
				var readOp api.ReadOp
				if err := json.Unmarshal(msg.Data, &readOp); err != nil {
					panic(err)
				}

				// Call stfs functions using the right parameters here and send answer structs
				n, err := os.Stderr.Write([]byte("stfs"))

				msg, err := json.Marshal(api.NewReadOpResponse(n, err.Error()))
				if err != nil {
					panic(err)
				}

				networking.WriteToDataChannel(msg)

				break
			case api.OpcodeWrite:
				var writeOp api.WriteOp
				if err := json.Unmarshal(msg.Data, &writeOp); err != nil {
					panic(err)
				}

				// Call stfs functions using the right parameters here and send answer structs
				n, err := os.Stderr.Write([]byte("stfs"))

				msg, err := json.Marshal(api.NewWriteOpResponse(int64(n), err.Error()))
				if err != nil {
					panic(err)
				}

				networking.WriteToDataChannel(msg)

				break
			case api.OpcodeSeek:
				var seekOp api.SeekOp
				if err := json.Unmarshal(msg.Data, &seekOp); err != nil {
					panic(err)
				}

				// Call stfs functions using the right parameters here and send answer structs
				n, err := os.Stderr.Seek(1, 2)

				msg, err := json.Marshal(api.NewSeekOpResponse(n, err.Error()))
				if err != nil {
					panic(err)
				}

				networking.WriteToDataChannel(msg)

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
	serverCmd.PersistentFlags().String(communityKey, "testCommunityName", "Community to join")

	viper.SetEnvPrefix("airdrip")
	viper.AutomaticEnv()
}

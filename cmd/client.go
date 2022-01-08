package cmd

import (
	"encoding/json"
	"fmt"
	"log"

	api "github.com/alphahorizonio/libentangle/pkg/api/websockets/v1"
	"github.com/alphahorizonio/libentangle/pkg/networking"
	"github.com/alphahorizonio/libentangle/pkg/readwriteseeker"
	"github.com/pion/webrtc/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start a signaling client.",
	RunE: func(cmd *cobra.Command, args []string) error {

		rmFile := readwriteseeker.RemoteFile{
			ReadCh:  make(chan api.ReadOpResponse),
			WriteCh: make(chan api.WriteOpResponse),
			SeekCh:  make(chan api.SeekOpResponse),
			OpenCh:  make(chan api.OpenOpResponse),
			CloseCh: make(chan api.CloseOpResponse),
		}

		networking.Connect("test", func(msg webrtc.DataChannelMessage) {

			fmt.Println(string(msg.Data))

			var v api.Message

			if err := json.Unmarshal(msg.Data, &v); err != nil {
				panic(err)
			}

			switch v.Opcode {
			case api.OpcodeOpenResponse:
				var openOpResponse api.OpenOpResponse
				if err := json.Unmarshal(msg.Data, &openOpResponse); err != nil {
					panic(err)
				}

				rmFile.OpenCh <- openOpResponse

				break
			case api.OpcodeCloseResponse:
				var closeOpResponse api.CloseOpResponse
				if err := json.Unmarshal(msg.Data, &closeOpResponse); err != nil {
					panic(err)
				}

				rmFile.CloseCh <- closeOpResponse

				break
			case api.OpcodeReadResponse:
				var readOpResponse api.ReadOpResponse
				if err := json.Unmarshal(msg.Data, &readOpResponse); err != nil {
					panic(err)
				}

				rmFile.ReadCh <- readOpResponse

				break
			case api.OpcodeWriteResponse:
				var writeOpResponse api.WriteOpResponse
				if err := json.Unmarshal(msg.Data, &writeOpResponse); err != nil {
					panic(err)
				}

				rmFile.WriteCh <- writeOpResponse

				break
			case api.OpcodeSeekResponse:
				var seekOpResponse api.SeekOpResponse
				if err := json.Unmarshal(msg.Data, &seekOpResponse); err != nil {
					panic(err)
				}

				rmFile.SeekCh <- seekOpResponse

				break
			}
		})

		select {}
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

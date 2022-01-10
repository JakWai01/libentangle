package cmd

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

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
		var file *os.File

		onOpen := make(chan struct{})
		dir, err := os.MkdirTemp(os.TempDir(), "serverfiles-*")
		if err != nil {
			panic(err)
		}

		myFile := filepath.Join(dir, "serverfile.tar")

		networking.Connect("test", func(msg webrtc.DataChannelMessage) {

			log.Println(string(msg.Data))

			var v api.Message

			if err := json.Unmarshal(msg.Data, &v); err != nil {
				panic(err)
			}

			switch v.Opcode {
			case api.OpcodeOpen:
				var openOp api.OpenOp
				if err := json.Unmarshal(msg.Data, &openOp); err != nil {
					panic(err)
				}
				log.Println(openOp.ReadOnly)
				if openOp.ReadOnly == true {
					file, err = os.Open(myFile)
					if err != nil {
						msg, err := json.Marshal(api.NewOpenOpResponse(err.Error()))
						if err != nil {
							panic(err)
						}

						networking.WriteToDataChannel(msg)
					} else {
						msg, err := json.Marshal(api.NewOpenOpResponse(""))
						if err != nil {
							panic(err)
						}

						networking.WriteToDataChannel(msg)
					}
				} else {
					file, err = os.OpenFile(myFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
					if err != nil {
						msg, err := json.Marshal(api.NewOpenOpResponse(err.Error()))
						if err != nil {
							panic(err)
						}

						networking.WriteToDataChannel(msg)
					} else {
						msg, err := json.Marshal(api.NewOpenOpResponse(""))
						if err != nil {
							panic(err)
						}

						networking.WriteToDataChannel(msg)
					}
				}

				break
			case api.OpcodeClose:
				var closeOp api.CloseOp
				if err := json.Unmarshal(msg.Data, &closeOp); err != nil {
					panic(err)
				}

				file.Close()

				msg, err := json.Marshal(api.NewCloseOpResponse(""))
				if err != nil {
					panic(err)
				}

				networking.WriteToDataChannel(msg)

				break
			case api.OpcodeRead:
				var readOp api.ReadOp
				if err := json.Unmarshal(msg.Data, &readOp); err != nil {
					panic(err)
				}

				buf := make([]byte, readOp.Length)

				n, err := file.Read(buf)
				if err != nil {
					msg, err := json.Marshal(api.NewReadOpResponse(buf, int64(n), err.Error()))
					if err != nil {
						panic(err)
					}

					networking.WriteToDataChannel(msg)
				} else {
					msg, err := json.Marshal(api.NewReadOpResponse(buf, int64(n), ""))
					if err != nil {
						panic(err)
					}

					networking.WriteToDataChannel(msg)
				}

				break
			case api.OpcodeWrite:
				var writeOp api.WriteOp
				if err := json.Unmarshal(msg.Data, &writeOp); err != nil {
					panic(err)
				}

				n, err := file.Write(writeOp.Payload)
				if err != nil {
					msg, err := json.Marshal(api.NewWriteOpResponse(int64(n), err.Error()))
					if err != nil {
						panic(err)
					}

					networking.WriteToDataChannel(msg)
				} else {
					msg, err := json.Marshal(api.NewWriteOpResponse(int64(n), ""))
					if err != nil {
						panic(err)
					}

					networking.WriteToDataChannel(msg)
				}

				break
			case api.OpcodeSeek:
				var seekOp api.SeekOp
				if err := json.Unmarshal(msg.Data, &seekOp); err != nil {
					panic(err)
				}

				offset, err := file.Seek(seekOp.Offset, seekOp.Whence)
				if err != nil {
					msg, err := json.Marshal(api.NewSeekOpResponse(offset, err.Error()))
					if err != nil {
						panic(err)
					}

					networking.WriteToDataChannel(msg)
				} else {
					msg, err := json.Marshal(api.NewSeekOpResponse(offset, ""))
					if err != nil {
						panic(err)
					}

					networking.WriteToDataChannel(msg)
				}

				break
			}
		}, func() {
			onOpen <- struct{}{}
		})

		<-onOpen

		select {}
	},
}

func init() {
	serverCmd.PersistentFlags().String(communityKey, "testCommunityName", "Community to join")

	viper.SetEnvPrefix("airdrip")
	viper.AutomaticEnv()
}

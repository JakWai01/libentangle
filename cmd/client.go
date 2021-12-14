package cmd

import (
	"bufio"
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

		networking.Connect("test", func(msg webrtc.DataChannelMessage) { log.Printf("Message: %s", msg.Data) })

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

package cmd

import (
	"bufio"
	"log"
	"os"

	"github.com/alphahorizon/libentangle/pkg/networking"
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

		networking.Connect("test")

		for {
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			networking.Write([]byte(text))

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

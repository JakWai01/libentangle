package cmd

import (
	"bufio"
	"log"
	"os"
	"sync"

	"github.com/alphahorizon/libentangle/pkg/signaling"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"nhooyr.io/websocket"
)

const (
	communityKey = "community"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start a signaling client.",
	RunE: func(cmd *cobra.Command, args []string) error {

		manager := signaling.NewClientManager()

		client := signaling.NewSignalingClient(
			func(conn *websocket.Conn, uuid string) error {
				return manager.HandleAcceptance(conn, uuid)
			},
			func(conn *websocket.Conn, data []byte, uuid string, wg *sync.WaitGroup) error {
				return manager.HandleIntroduction(conn, data, uuid, wg)
			},
			func(conn *websocket.Conn, data []byte, candidates *chan string, wg *sync.WaitGroup, uuid string) error {
				return manager.HandleOffer(conn, data, candidates, wg, uuid)
			},
			func(data []byte, candidates *chan string, wg *sync.WaitGroup) error {
				return manager.HandleAnswer(data, candidates, wg)
			},
			func(data []byte, candidates *chan string) error {
				return manager.HandleCandidate(data, candidates)
			},
			func() error {
				return manager.HandleResignation()
			},
		)

		go func() {
			go client.HandleConn("localhost:9090", viper.GetString(communityKey))
		}()

		// Use generic send function to send a message, after the connection was established
		for {
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			manager.SendMessage(text)

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

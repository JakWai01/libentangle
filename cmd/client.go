package cmd

import (
	"log"
	"os"
	"sync"

	"github.com/alphahorizon/libentangle/pkg/signaling"
	"github.com/pion/webrtc/v3"
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

		fatal := make(chan error)
		done := make(chan struct{})

		manager := signaling.NewClientManager()

		client := signaling.NewSignalingClient(
			func(conn *websocket.Conn, uuid string) error {
				return manager.HandleAcceptance(conn, uuid)
			},
			func(conn *websocket.Conn, data []byte, peerConnection *webrtc.PeerConnection, uuid string) error {
				return manager.HandleIntroduction(conn, data, peerConnection, uuid)
			},
			func(conn *websocket.Conn, data []byte, peerConnection *webrtc.PeerConnection, candidates *chan string, wg *sync.WaitGroup, uuid string) error {
				return manager.HandleOffer(conn, data, peerConnection, candidates, wg, uuid)
			},
			func(data []byte, peerConnection *webrtc.PeerConnection, candidates *chan string, wg *sync.WaitGroup) error {
				return manager.HandleAnswer(data, peerConnection, candidates, wg)
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

		for {
			select {
			case err := <-fatal:
				log.Fatal(err)
			case <-done:
				os.Exit(0)
			}
		}
	},
}

func init() {
	clientCmd.PersistentFlags().String(communityKey, "a", "Community to join")

	// Bind env variables
	if err := viper.BindPFlags(clientCmd.PersistentFlags()); err != nil {
		log.Fatal("could not bind flags:", err)
	}
	viper.SetEnvPrefix("airdrip")
	viper.AutomaticEnv()
}

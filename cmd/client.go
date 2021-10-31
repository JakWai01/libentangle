package cmd

import (
	"log"
	"os"

	"github.com/alphahorizon/libentangle/pkg/signaling"
	"github.com/spf13/cobra"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start a signaling client.",
	RunE: func(cmd *cobra.Command, args []string) error {

		fatal := make(chan error)
		done := make(chan struct{})

		client := signaling.NewSignalingClient()

		go func() {
			go client.HandleConn("localhost:9090", "test", "", []byte(""))
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

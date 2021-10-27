package cmd

import (
	"log"
	"net"
	"net/http"

	"github.com/alphahorizon/libentangle/pkg/signaling"
	"github.com/spf13/cobra"
	"nhooyr.io/websocket"
)

var signalCmd = &cobra.Command{
	Use:   "signal",
	Short: "Start a signaling server",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle lifecycle
		signaler := signaling.NewSignalingServer()

		for {
			addr, err := net.ResolveTCPAddr("tcp", "localhost:9090")
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("signaling server listening on %v", addr)

			handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				conn, err := websocket.Accept(rw, r, &websocket.AcceptOptions{
					InsecureSkipVerify: true, // CORS
				})
				if err != nil {
					log.Fatal(err)
				}

				log.Println("client connected")

				go func() {
					signaler.HandleConn(*conn)
				}()
			})

			http.ListenAndServe(addr.String(), handler)
		}
	},
}

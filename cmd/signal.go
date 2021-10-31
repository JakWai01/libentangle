package cmd

import (
	"log"
	"net"
	"net/http"

	api "github.com/alphahorizon/libentangle/pkg/api/websockets/v1"
	"github.com/alphahorizon/libentangle/pkg/signaling"
	"github.com/spf13/cobra"
	"nhooyr.io/websocket"
)

var signalCmd = &cobra.Command{
	Use:   "signal",
	Short: "Start a signaling server",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle lifecycle

		for {
			addr, err := net.ResolveTCPAddr("tcp", "localhost:9090")
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("signaling server listening on %v", addr)

			communities := signaling.NewCommunitiesManager()

			signaler := signaling.NewSignalingServer(
				func(application api.Application, conn *websocket.Conn) error {
					return communities.HandleApplication(application, conn)
				},
				func(ready api.Ready, conn *websocket.Conn) error {
					return communities.HandleReady(ready, conn)
				},
				func(offer api.Offer) error {
					return communities.HandleOffer(offer)
				},
				func(answer api.Answer) error {
					return communities.HandleAnswer(answer)
				},
				func(candidate api.Candidate) error {
					return communities.HandleCandidate(candidate)
				},
				func(exited api.Exited) error {
					return communities.HandleExited(exited)
				},
			)

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
